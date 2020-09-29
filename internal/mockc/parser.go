package mockc

import (
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

type parser struct {
	pkg *packages.Package
}

func newParser(pkg *packages.Package) *parser {
	return &parser{
		pkg: pkg,
	}
}

func (p *parser) Parse() ([]*generator, error) {
	destinationsAndGenerators := map[string]*generator{}
	for _, syntax := range p.pkg.Syntax {
		for _, decl := range syntax.Decls {
			fun, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			calls, err := p.findMockcCalls(fun.Body.List)
			if err != nil {
				errorMessage := "cannot find mockc calls:"
				errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, err)

				return nil, errors.New(errorMessage)
			} else if len(calls) == 0 {
				continue
			}

			var (
				pkgDir          = filepath.Dir(p.pkg.Fset.File(decl.Pos()).Name())
				destination     = defaultDestination
				mockName        = fun.Name.Name
				hasConstructor  bool
				fieldNamePrefix = defaultFieldNamePrefix
				fieldNameSuffix = defaultFieldNameSuffix
				interfaces      []*types.Interface
			)

			for _, call := range calls {
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				obj := p.pkg.TypesInfo.ObjectOf(sel.Sel)
				switch obj.Name() {
				case "Implements":
					log.Println("mockc.Implements is deprecated. Please use mock.Implement instead.")
					fallthrough
				case "Implement":
					for _, arg := range call.Args {
						t := p.pkg.TypesInfo.TypeOf(arg)

						inter, ok := t.Underlying().(*types.Interface)
						if !ok {
							errorMessage := "non-interface:"
							errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, t)

							return nil, errors.New(errorMessage)
						}

						var isExternalInterface bool
						switch arg := arg.(*ast.CallExpr).Fun.(type) {
						case *ast.SelectorExpr:
							isExternalInterface = p.pkg.TypesInfo.ObjectOf(arg.Sel).Pkg() != p.pkg.Types
						case *ast.Ident:
							isExternalInterface = p.pkg.TypesInfo.ObjectOf(arg).Pkg() != p.pkg.Types
						case *ast.InterfaceType:
						default:
							errorMessage := "unknown interface:"
							errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, t)

							return nil, errors.New(errorMessage)
						}

						err := validateInterface(inter, isExternalInterface)
						if err != nil {
							errorMessage := "invalid interface:"
							errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, err)

							return nil, errors.New(errorMessage)
						}

						interfaces = append(interfaces, inter)
					}
				case "SetFieldNamePrefix":
					arg := call.Args[0]
					res, err := types.Eval(p.pkg.Fset, p.pkg.Types, arg.Pos(), types.ExprString(arg))
					if err != nil {
						errorMessage := "cannot set field name prefix:"
						errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, err)

						return nil, errors.New(errorMessage)
					}

					val := res.Value.ExactString()

					fieldNamePrefix = val[1 : len(val)-1]
				case "SetFieldNameSuffix":
					arg := call.Args[0]
					res, err := types.Eval(p.pkg.Fset, p.pkg.Types, arg.Pos(), types.ExprString(arg))
					if err != nil {
						errorMessage := "cannot set field name suffix:"
						errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, err)

						return nil, errors.New(errorMessage)
					}

					val := res.Value.ExactString()

					fieldNameSuffix = val[1 : len(val)-1]
				case "SetDestination":
					arg := call.Args[0]
					res, err := types.Eval(p.pkg.Fset, p.pkg.Types, arg.Pos(), types.ExprString(arg))
					if err != nil {
						errorMessage := "cannot set destination:"
						errorMessage += fmt.Sprintf("\n\tmock %q: %v", fun.Name.Name, err)

						return nil, errors.New(errorMessage)
					}

					val := res.Value.ExactString()
					val = val[1 : len(val)-1]

					if val == "" {
						errorMessage := "cannot set destination:"
						errorMessage += fmt.Sprintf("\n\tmock %q: destination should not be an empty string", fun.Name.Name)

						return nil, errors.New(errorMessage)
					}

					if filepath.Ext(val) != ".go" {
						errorMessage := "cannot set destination:"
						errorMessage += fmt.Sprintf("\n\tmock %q: %q is not a go file", fun.Name.Name, val)

						return nil, errors.New(errorMessage)
					}

					destination = val
				case "WithConstructor":
					hasConstructor = true
				default:
					errorMessage := "unknown mockc function call:"
					errorMessage += fmt.Sprintf("\n\tmock %q: mockc.%s", fun.Name.Name, obj.Name())

					return nil, errors.New(errorMessage)
				}
			}

			if fieldNamePrefix == "" && fieldNameSuffix == "" {
				errorMessage := "at least one of the field name prefix and field name suffix must not be an empty string:"
				errorMessage += fmt.Sprintf("\n\tmock %q: prefix(%q) suffix(%q)", fun.Name.Name, fieldNamePrefix, fieldNameSuffix)

				return nil, errors.New(errorMessage)
			}

			destination = filepath.Join(pkgDir, destination)
			if destinationsAndGenerators[destination] == nil {
				destinationsAndGenerators[destination] = newGenerator(p.pkg, destination)
			}
			g := destinationsAndGenerators[destination]

			err = g.addMock(mockName, hasConstructor, newFieldNameFormatter(fieldNamePrefix, fieldNameSuffix), interfaces)
			if err != nil {
				return nil, err
			}
		}
	}
	if len(destinationsAndGenerators) == 0 {
		return nil, nil
	}

	generators := make([]*generator, 0, len(destinationsAndGenerators))
	for _, generator := range destinationsAndGenerators {
		generators = append(generators, generator)
	}

	return generators, nil
}

func (p *parser) findMockcCalls(stmts []ast.Stmt) ([]*ast.CallExpr, error) {
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

			if p.pkg.TypesInfo.ObjectOf(sel.Sel).Pkg().Path() == mockcPath {
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
		return nil, errors.New("mock generator should be consist of mockc function calls")
	}

	return calls, nil
}
