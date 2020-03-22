package mockc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

type generator struct {
	pkg             *packages.Package
	path            string
	imports         map[string]string
	importConflicts map[string]int
	mocks           []mockInfo
}

func newGenerator(pkg *packages.Package, path string) *generator {
	return &generator{
		pkg:             pkg,
		path:            path,
		imports:         map[string]string{},
		importConflicts: map[string]int{},
	}
}

func (g *generator) loadMocks() error {
	for _, syntax := range g.pkg.Syntax {
		for _, decl := range syntax.Decls {
			fun, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			calls, err := g.findMockcCalls(fun.Body.List)
			if err != nil {
				return err
			} else if len(calls) == 0 {
				continue
			}

			var (
				mockName        = fun.Name.Name
				fieldNamePrefix = "_"
				fieldNameSuffix = ""
				interfaces      []*types.Interface
			)

			for _, call := range calls {
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				obj := g.pkg.TypesInfo.ObjectOf(sel.Sel)
				switch obj.Name() {
				case "Implements":
					log.Println("mockc.Implements is deprecated. Please use mock.Implement instead.")
					fallthrough
				case "Implement":
					for _, arg := range call.Args {
						t := g.pkg.TypesInfo.TypeOf(arg)

						inter, ok := t.Underlying().(*types.Interface)
						if !ok {
							errorMessage := "non-interface:"
							errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, t)

							return errors.New(errorMessage)
						}

						var isExternalInterface bool
						switch arg := arg.(*ast.CallExpr).Fun.(type) {
						case *ast.SelectorExpr:
							isExternalInterface = g.pkg.TypesInfo.ObjectOf(arg.Sel).Pkg() != g.pkg.Types
						case *ast.Ident:
							isExternalInterface = g.pkg.TypesInfo.ObjectOf(arg).Pkg() != g.pkg.Types
						case *ast.InterfaceType:
						default:
							errorMessage := "unknown interface:"
							errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, t)

							return errors.New(errorMessage)
						}

						err := g.validateInterface(inter, isExternalInterface)
						if err != nil {
							errorMessage := "invalid interface:"
							errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, err)

							return errors.New(errorMessage)
						}

						interfaces = append(interfaces, inter)
					}
				case "SetFieldNamePrefix":
					arg := call.Args[0]
					res, err := types.Eval(g.pkg.Fset, g.pkg.Types, arg.Pos(), types.ExprString(arg))
					if err != nil {
						errorMessage := "cannot set field name prefix:"
						errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, err)

						return errors.New(errorMessage)
					}

					val := res.Value.ExactString()

					fieldNamePrefix = val[1 : len(val)-1]
				case "SetFieldNameSuffix":
					arg := call.Args[0]
					res, err := types.Eval(g.pkg.Fset, g.pkg.Types, arg.Pos(), types.ExprString(arg))
					if err != nil {
						errorMessage := "cannot set field name suffix:"
						errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, err)

						return errors.New(errorMessage)
					}

					val := res.Value.ExactString()

					fieldNameSuffix = val[1 : len(val)-1]
				default:
					errorMessage := "unknown mockc function call:"
					errorMessage += fmt.Sprintf("\n\tmock %q: mockc.%s", fun.Name.Name, obj.Name())

					return errors.New(errorMessage)
				}
			}

			if fieldNamePrefix == "" && fieldNameSuffix == "" {
				errorMessage := "at least one of the field name prefix and field name suffix must not be an empty string:"
				errorMessage += fmt.Sprintf("\n\tmock %q: prefix(%q) suffix(%q)", fun.Name.Name, fieldNamePrefix, fieldNameSuffix)

				return errors.New(errorMessage)
			}

			mock, err := g.newMock(mockName, newFieldNameFormatter(fieldNamePrefix, fieldNameSuffix), interfaces)
			if err != nil {
				return err
			}

			g.mocks = append(g.mocks, mock)
		}
	}

	return nil
}

func (g *generator) loadMockWithFlags(ctx context.Context, wd string, name string, fieldNamePrefix string, fieldNameSuffix string, interfacePatterns []string) error {
	targetInterfaces := map[string][]string{}
	for _, inter := range interfacePatterns {
		idx := strings.LastIndex(inter, ".")
		if idx == -1 {
			errorMessage := "invalid interface pattern:"
			errorMessage += fmt.Sprintf("\n\texpected interface pattern {package_path}.{interface_name}: actual %s", inter)

			return errors.New(errorMessage)
		}

		pkgPath, interfaceName := inter[:idx], inter[idx+1:]
		if pkgPath == "" || interfaceName == "" {
			errorMessage := "invalid interface pattern:"
			errorMessage += fmt.Sprintf("\n\texpected interface pattern {package-path}.{interface-name}: actual %s", inter)

			return errors.New(errorMessage)
		}

		targetInterfaces[pkgPath] = append(targetInterfaces[pkgPath], interfaceName)
	}

	patterns := make([]string, 0, len(targetInterfaces))
	for pkgPath, _ := range targetInterfaces {
		patterns = append(patterns, pkgPath)
	}

	pkgs, err := loadPackages(ctx, wd, patterns)
	if err != nil {
		return fmt.Errorf("cannot load packages: %v", err)
	}

	interfaces := make([]*types.Interface, 0, len(interfacePatterns))
	for _, pkg := range pkgs {
		interfaceNames := targetInterfaces[pkg.PkgPath]
		if len(interfaceNames) == 0 {
			continue
		}

		f := newInterfaceFinder(pkg, interfaceNames)
		for _, syntax := range pkg.Syntax {
			ast.Walk(f, syntax)
		}

		for _, interfaceName := range interfaceNames {
			inter, ok := f.result[interfaceName]
			if !ok {
				return fmt.Errorf("\n\tpackage %q: cannot load interface: %s", pkg.PkgPath, interfaceName)
			}

			err = g.validateInterface(inter, g.pkg.PkgPath != pkg.PkgPath)
			if err != nil {
				return fmt.Errorf("\n\tpackage %q: invalid interface: %v", pkg.PkgPath, err)
			}

			interfaces = append(interfaces, inter)
		}
	}

	mock, err := g.newMock(name, newFieldNameFormatter(fieldNamePrefix, fieldNameSuffix), interfaces)
	if err != nil {
		return fmt.Errorf("cannot create mock: %v", err)
	}

	g.mocks = []mockInfo{mock}

	return nil
}

func (g *generator) generate(gogenerate string) error {
	if len(g.mocks) == 0 {
		return nil
	}

	b := bytes.NewBuffer(nil)

	err := tmpl.Execute(b, struct {
		PackageName string
		GoGenerate  string
		Imports     map[string]string
		Mocks       []mockInfo
	}{
		PackageName: g.pkg.Name,
		GoGenerate:  gogenerate,
		Imports:     g.imports,
		Mocks:       g.mocks,
	})
	if err != nil {
		return fmt.Errorf("cannot execute template: %v", err)
	}

	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return fmt.Errorf("cannot format mockc generated code: %v", err)
	}

	f, err := os.Create(g.path)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("cannot create %s: %v", g.path, err)
	}

	_, err = f.Write(formatted)
	if err != nil {
		return fmt.Errorf("cannot write %s: %v", g.path, err)
	}

	log.Println("generated:", g.path)

	return nil
}

func (g *generator) newMock(mockName string, fieldNameFormatter func(string) string, interfaces []*types.Interface) (mockInfo, error) {
	mock := mockInfo{
		Name: mockName,
	}

	funs := map[string]*types.Func{}
	for _, inter := range interfaces {
		for i := 0; i < inter.NumMethods(); i++ {
			fun := inter.Method(i)
			if f, ok := funs[fun.Name()]; ok && fun.Type().(*types.Signature).String() != f.Type().(*types.Signature).String() {
				errorMessage := "duplicated method:"
				errorMessage += fmt.Sprintf("\n\tmock %q: method %q", mock.Name, fun.Name())

				return mockInfo{}, errors.New(errorMessage)
			}

			funs[fun.Name()] = fun
		}
	}

	mock.Methods = make([]methodInfo, 0, len(funs))
	for funName, fun := range funs {
		methodInfo := methodInfo{
			Name:      funName,
			FieldName: fieldNameFormatter(funName),
		}

		sig := fun.Type().(*types.Signature)

		methodInfo.Params = make([]paramInfo, 0, sig.Params().Len())
		for i := 0; i < sig.Params().Len(); i++ {
			param := sig.Params().At(i)

			name := fmt.Sprintf("P%d", i)
			paramName := fmt.Sprintf("p%d", i)

			methodInfo.Params = append(methodInfo.Params, paramInfo{
				name:       name,
				paramName:  paramName,
				typeString: g.typeString(param.Type()),
				isVariadic: i+1 == sig.Params().Len() && sig.Variadic(),
			})
		}

		methodInfo.Results = make([]resultInfo, 0, sig.Results().Len())
		for i := 0; i < sig.Results().Len(); i++ {
			result := sig.Results().At(i)

			name := fmt.Sprintf("R%d", i)
			resultName := fmt.Sprintf("r%d", i)

			methodInfo.Results = append(methodInfo.Results, resultInfo{
				name:       name,
				resultName: resultName,
				typeString: g.typeString(result.Type()),
			})
		}

		mock.Methods = append(mock.Methods, methodInfo)
	}

	return mock, nil
}

func (g *generator) findMockcCalls(stmts []ast.Stmt) ([]*ast.CallExpr, error) {
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

			if g.pkg.TypesInfo.ObjectOf(sel.Sel).Pkg().Path() == mockcPath {
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

func (g *generator) validateInterface(inter *types.Interface, isExternalInterface bool) error {
	for i := 0; i < inter.NumMethods(); i++ {
		method := inter.Method(i)
		if isExternalInterface && !method.Exported() {
			return fmt.Errorf("cannot implement non-exported method: %s", method.FullName())
		}
	}

	return nil
}

func (g *generator) typeString(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		return t.Name()
	case *types.Pointer:
		return "*" + g.typeString(t.Elem())
	case *types.Slice:
		return "[]" + g.typeString(t.Elem())
	case *types.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), g.typeString(t.Elem()))
	case *types.Map:
		kt := g.typeString(t.Key())
		vt := g.typeString(t.Elem())

		return fmt.Sprintf("map[%s]%s", kt, vt)
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			return "chan " + g.typeString(t.Elem())
		case types.RecvOnly:
			return "<-chan " + g.typeString(t.Elem())
		default:
			return "chan<- " + g.typeString(t.Elem())
		}
	case *types.Struct:
		var fields []string

		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)

			if f.Anonymous() {
				fields = append(fields, g.typeString(f.Type()))
			} else {
				fields = append(fields, fmt.Sprintf("%s %s", f.Name(), g.typeString(f.Type())))
			}
		}

		return fmt.Sprintf("struct{%s}", strings.Join(fields, ";"))
	case *types.Named:
		o := t.Obj()
		if o.Pkg() == nil || o.Pkg().Path() == g.pkg.PkgPath {
			return o.Name()
		}

		return g.getUniquePackageName(o.Pkg().Path(), o.Pkg().Name()) + "." + o.Name()
	case *types.Signature:
		switch t.Results().Len() {
		case 0:
			return fmt.Sprintf(
				"func(%s)",
				g.tupleTypeString(t.Params(), t.Variadic()),
			)
		case 1:
			return fmt.Sprintf(
				"func(%s) %s",
				g.tupleTypeString(t.Params(), t.Variadic()),
				g.typeString(t.Results().At(0).Type()),
			)
		default:
			return fmt.Sprintf(
				"func(%s)(%s)",
				g.tupleTypeString(t.Params(), t.Variadic()),
				g.tupleTypeString(t.Results(), false),
			)
		}
	case *types.Interface:
		methods := make([]string, 0, t.NumMethods())
		for i := 0; i < t.NumMethods(); i++ {
			method := t.Method(i)
			sig := method.Type().(*types.Signature)

			switch sig.Results().Len() {
			case 0:
				methods = append(methods, fmt.Sprintf(
					"%s(%s)",
					method.Name(),
					g.tupleTypeString(sig.Params(), sig.Variadic()),
				))
			case 1:
				methods = append(methods, fmt.Sprintf(
					"%s(%s) %s",
					method.Name(),
					g.tupleTypeString(sig.Params(), sig.Variadic()),
					g.typeString(sig.Results().At(0).Type()),
				))
			default:
				methods = append(methods, fmt.Sprintf(
					"%s(%s) (%s)",
					method.Name(),
					g.tupleTypeString(sig.Params(), sig.Variadic()),
					g.tupleTypeString(sig.Results(), false),
				))
			}
		}

		return fmt.Sprintf("interface{%s}", strings.Join(methods, ";"))
	default:
		return ""
	}
}

func (g *generator) tupleTypeString(t *types.Tuple, variadic bool) string {
	var typeStrings []string
	for i := 0; i < t.Len(); i++ {
		typeString := g.typeString(t.At(i).Type())
		if variadic {
			typeString = "..." + typeString[2:]
		}

		typeStrings = append(typeStrings, typeString)
	}

	return strings.Join(typeStrings, ", ")
}

func (g *generator) getUniquePackageName(path string, name string) string {
	if uname, ok := g.imports[path]; ok {
		return uname
	}

	uname := name
	cnt := g.importConflicts[uname]
	g.importConflicts[uname]++
	if cnt != 0 {
		uname += strconv.Itoa(cnt)
	}

	g.imports[path] = uname

	return uname
}

func newFieldNameFormatter(prefix, suffix string) func(string) string {
	return func(field string) string {
		return prefix + field + suffix
	}
}
