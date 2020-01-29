package mockc

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

type Package struct {
	underlying *packages.Package

	packageNames  map[string]string
	conflictNames map[string]int

	mocks []Mock
}

func newPackage(pkg *packages.Package) (*Package, error) {
	p := &Package{
		underlying:    pkg,
		packageNames:  map[string]string{},
		conflictNames: map[string]int{},
	}

	if !p.hasMockc() {
		return p, nil
	}

	err := p.loadMocks()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Package) hasMockc() bool {
	for _, imps := range p.underlying.Imports {
		if imps.PkgPath == packageName {
			return true
		}
	}

	return false
}

func (p *Package) loadMocks() error {
	for _, syntax := range p.underlying.Syntax {
		for _, decl := range syntax.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			calls, err := p.findMockcCalls(fn.Body.List)
			if err != nil {
				return err
			} else if len(calls) == 0 {
				continue
			}

			mock := Mock{
				Name: fn.Name.Name,
			}

			for _, call := range calls {
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				obj := p.underlying.TypesInfo.ObjectOf(sel.Sel)
				switch obj.Name() {
				case "Implements":
					for _, arg := range call.Args {
						inter, err := p.getValidInterface(arg)
						if err != nil {
							errorMessage := "invalid interface:"
							errorMessage += fmt.Sprintf("\n\t%v: func %v()", p.underlying.PkgPath, fn.Name.Name)
							errorMessage += fmt.Sprintf("\n\terr: %v", err)

							return errors.New(errorMessage)
						}

						mock.Interfaces = append(mock.Interfaces, inter)
					}
				case "SetMode":
					if mock.Mode != 0 {
						errorMessage := "only one mockc.SetMode can be in the function:"
						errorMessage += fmt.Sprintf("\n\t%v: func %v()", p.underlying.PkgPath, fn.Name.Name)

						return errors.New(errorMessage)
					}

					mock.Mode, err = p.evalMode(call.Args[0])
					if err != nil {
						errorMessage := "cannot get mockc.Mode:"
						errorMessage += fmt.Sprintf("\n\t%v: func %v()", p.underlying.PkgPath, fn.Name.Name)
						errorMessage += fmt.Sprintf("\n\terr: %v", err)

						return errors.New(errorMessage)
					}
				default:
					errorMessage := "unknown mockc function:"
					errorMessage += fmt.Sprintf("\n\t%v: func %v()", p.underlying.PkgPath, fn.Name.Name)
					errorMessage += fmt.Sprintf("\n\tfunc: %v", obj.Name())

					return errors.New(errorMessage)
				}
			}

			p.mocks = append(p.mocks, mock)
		}
	}

	return nil
}

func (p *Package) findMockcCalls(stmts []ast.Stmt) ([]*ast.CallExpr, error) {
	var calls []*ast.CallExpr
	var invalid bool

	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *ast.ExprStmt:
			call, ok := stmt.X.(*ast.CallExpr)
			if !ok {
				continue
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				continue
			}

			if p.underlying.TypesInfo.ObjectOf(sel.Sel).Pkg().Path() == packageName {
				calls = append(calls, call)
			}
		case *ast.EmptyStmt, *ast.ReturnStmt:
		default:
			invalid = true
		}
	}

	if len(calls) == 0 {
		return nil, nil
	}

	if invalid {
		return nil, errors.New("mockc generator should be consist of mockc function calls")
	}

	return calls, nil
}

func (p *Package) getValidInterface(arg ast.Expr) (*types.Interface, error) {
	t := p.underlying.TypesInfo.TypeOf(arg)

	inter, ok := t.Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("'%v' is not a interface", t)
	}

	var external bool
	switch arg := arg.(*ast.CallExpr).Fun.(type) {
	case *ast.SelectorExpr:
		external = p.underlying.TypesInfo.ObjectOf(arg.Sel).Pkg() != p.underlying.Types
	case *ast.Ident:
		external = p.underlying.TypesInfo.ObjectOf(arg).Pkg() != p.underlying.Types
	case *ast.InterfaceType:
		// Do nothing.
	default:
		return nil, fmt.Errorf("unknown interface: %v", t)
	}

	for i := 0; i < inter.NumMethods(); i++ {
		method := inter.Method(i)
		if external && !method.Exported() {
			return nil, fmt.Errorf("cannot implement non-exported method: %v", method.FullName())
		}
	}

	return inter, nil
}

func (p *Package) evalMode(expr ast.Expr) (Mode, error) {
	v, err := types.Eval(p.underlying.Fset, p.underlying.Types, expr.Pos(), types.ExprString(expr))
	if err != nil {
		return 0, err
	}

	mode, err := strconv.Atoi(v.Value.ExactString())
	if err != nil {
		return 0, err
	}

	return Mode(mode), nil
}

func (p *Package) GenerateResult() error {
	mockInfos := make([]mockInfo, 0, len(p.mocks))
	for _, mock := range p.mocks {
		info := mockInfo{
			Name: mock.Name,
		}

		methods := map[string]*types.Func{}
		for _, inter := range mock.Interfaces {
			for i := 0; i < inter.NumMethods(); i++ {
				fun := inter.Method(i)
				if f, ok := methods[fun.Name()]; ok && fun.Type().(*types.Signature).String() != f.Type().(*types.Signature).String() {
					errorMessage := "duplicated method"
					errorMessage += fmt.Sprintf("\n\t%v: func %v()", p.underlying.PkgPath, mock.Name)
					errorMessage += fmt.Sprintf("\n\tmethod: %v", fun.String())

					return errors.New(errorMessage)
				}

				methods[fun.Name()] = fun
			}
		}

		info.Methods = make([]methodInfo, 0, len(methods))
		for methodName, method := range methods {
			methodInfo := methodInfo{
				Name: methodName,
			}

			sig := method.Type().(*types.Signature)

			methodInfo.Params = make([]paramInfo, 0, sig.Params().Len())
			for i := 0; i < sig.Params().Len(); i++ {
				param := sig.Params().At(i)

				paramName := fmt.Sprintf("p%d", i)
				name := param.Name()
				if name == "" || name == "_" {
					name = paramName
				}

				methodInfo.Params = append(methodInfo.Params, paramInfo{
					name:       name,
					paramName:  paramName,
					typeString: p.typeString(param.Type()),
					isVariadic: i == sig.Params().Len()-1 && sig.Variadic(),
				})
			}

			methodInfo.Results = make([]resultInfo, 0, sig.Results().Len())
			for i := 0; i < sig.Results().Len(); i++ {
				result := sig.Results().At(i)

				resultName := fmt.Sprintf("r%d", i)
				name := result.Name()
				if name == "" || name == "_" {
					name = resultName
				}

				methodInfo.Results = append(methodInfo.Results, resultInfo{
					name:       name,
					resultName: resultName,
					typeString: p.typeString(result.Type()),
				})
			}

			info.Methods = append(info.Methods, methodInfo)
		}

		mockInfos = append(mockInfos, info)
	}

	path := filepath.Join(filepath.Dir(p.underlying.GoFiles[0]), "mockc_gen.go")
	b := bytes.NewBuffer(nil)

	err := mockTmpl.Execute(b, struct {
		PackageName string
		Imports     map[string]string
		Mocks       []mockInfo
	}{
		PackageName: p.underlying.Name,
		Imports:     p.packageNames,
		Mocks:       mockInfos,
	})
	if err != nil {
		return fmt.Errorf("%v: %v", p.underlying.Name, err)
	}

	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return fmt.Errorf("cannot format mockc generated codes: %v", err)
	}

	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("cannot create %v: %v", path, err)
	}

	_, err = f.Write(formatted)
	if err != nil {
		return fmt.Errorf("cannot write %v: %v", path, err)
	}

	fmt.Printf("%v\n", path)

	return nil
}

func (p *Package) typeString(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		return t.Name()
	case *types.Pointer:
		return "*" + p.typeString(t.Elem())
	case *types.Slice:
		return "[]" + p.typeString(t.Elem())
	case *types.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), p.typeString(t.Elem()))
	case *types.Map:
		kt := p.typeString(t.Key())
		vt := p.typeString(t.Elem())

		return fmt.Sprintf("map[%s]%s", kt, vt)
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			return "chan " + p.typeString(t.Elem())
		case types.RecvOnly:
			return "<-chan " + p.typeString(t.Elem())
		default:
			return "chan<- " + p.typeString(t.Elem())
		}
	case *types.Struct:
		var fields []string

		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)

			if f.Anonymous() {
				fields = append(fields, p.typeString(f.Type()))
			} else {
				fields = append(fields, fmt.Sprintf("%s %s", f.Name(), p.typeString(f.Type())))
			}
		}

		return fmt.Sprintf("struct{%s}", strings.Join(fields, ";"))
	case *types.Named:
		o := t.Obj()
		if o.Pkg() == nil {
			return o.Name()
		}

		return p.getUniquePackageName(o.Pkg().Path(), o.Pkg().Name()) + "." + o.Name()
	case *types.Signature:
		switch t.Results().Len() {
		case 0:
			return fmt.Sprintf(
				"func(%s)",
				p.tupleTypeString(t.Params()),
			)
		case 1:
			return fmt.Sprintf(
				"func(%s) %s",
				p.tupleTypeString(t.Params()),
				p.typeString(t.Results().At(0).Type()),
			)
		default:
			return fmt.Sprintf(
				"func(%s)(%s)",
				p.tupleTypeString(t.Params()),
				p.tupleTypeString(t.Results()),
			)
		}
	case *types.Interface:
		methods := make([]string, 0, t.NumMethods())
		for i := 0; i < t.NumMethods(); i++ {
			method := t.Method(i)
			sig := method.Type().(*types.Signature)

			switch sig.Results().Len() {
			case 0:
				methods = append(methods, fmt.Sprintf("%v(%v)", method.Name(), p.tupleTypeString(sig.Params())))
			case 1:
				methods = append(methods, fmt.Sprintf("%v(%v) %v", method.Name(), p.tupleTypeString(sig.Params()), p.typeString(sig.Results().At(0).Type())))
			default:
				methods = append(methods, fmt.Sprintf("%v(%v) (%v)", method.Name(), p.tupleTypeString(sig.Params()), p.tupleTypeString(sig.Results())))
			}
		}

		return fmt.Sprintf("interface{%v}", strings.Join(methods, ";"))
	default:
		return ""
	}
}

func (p *Package) tupleTypeString(t *types.Tuple) string {
	var typeStrings []string
	for i := 0; i < t.Len(); i++ {
		v := t.At(i)

		typeStrings = append(typeStrings, p.typeString(v.Type()))
	}

	return strings.Join(typeStrings, ", ")
}

func (p *Package) getUniquePackageName(path string, name string) string {
	if uname, ok := p.packageNames[path]; ok {
		return uname
	}

	uname := name
	cnt := p.conflictNames[uname]
	p.conflictNames[uname]++
	if cnt != 0 {
		uname += strconv.Itoa(cnt)
	}

	p.packageNames[path] = uname

	return uname
}
