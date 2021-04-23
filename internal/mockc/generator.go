package mockc

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"io/ioutil"
	"log"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/dave/jennifer/jen"
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

func (g *generator) Generate(gogenerate string) error {
	if len(g.mocks) == 0 {
		return nil
	}

	g.sortMocks()

	b, err := render(g.pkg, g.mocks, gogenerate)
	if err != nil {
		return fmt.Errorf("cannot execute template: %v", err)
	}

	err = ioutil.WriteFile(g.path, b, 0666)
	if err != nil {
		return fmt.Errorf("cannot write %s: %v", g.path, err)
	}

	log.Println("generated:", g.path)

	return nil
}

func (g *generator) sortMocks() {
	sort.Slice(g.mocks, func(i, j int) bool {
		return g.mocks[i].name < g.mocks[j].name
	})
	for _, m := range g.mocks {
		sort.Slice(m.methods, func(i, j int) bool {
			return m.methods[i].typ.Name() < m.methods[j].typ.Name()
		})
	}
}

func (g *generator) addMockWithFlags(ctx context.Context, wd string, name string, withConstructor bool, fieldNamePrefix string, fieldNameSuffix string, interfacePatterns []string) error {
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

	interfaces := make([]types.Type, 0, len(interfacePatterns))
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
				return fmt.Errorf("package %q: cannot load interface: %s", pkg.PkgPath, interfaceName)
			}

			interfaces = append(interfaces, inter)
		}
	}

	var constructor string
	if withConstructor {
		constructor = "New" + name
	}

	err = g.addMock(name, constructor, interfaces, newFieldNameFormatter(fieldNamePrefix, fieldNameSuffix))
	if err != nil {
		return err
	}

	return nil
}

func (g *generator) addMock(name string, constructor string, interfaces []types.Type, fieldNameFormatter func(string) string) error {
	mock := mockInfo{
		name:        name,
		constructor: constructor,
	}

	var (
		methods   []*types.Func
		embeddeds []types.Type
	)
	for _, inter := range interfaces {
		switch inter := inter.(type) {
		case *types.Named:
			embeddeds = append(embeddeds, inter)
		case *types.Interface:
			for i := 0; i < inter.NumEmbeddeds(); i++ {
				embeddeds = append(embeddeds, inter.EmbeddedType(i))
			}
			for i := 0; i < inter.NumExplicitMethods(); i++ {
				methods = append(methods, inter.ExplicitMethod(i))
			}
		}
	}

	mock.typ = types.NewInterfaceType(methods, embeddeds)
	err := complete(mock.typ)
	if err != nil {
		errorMessage := err.Error()
		errorMessage += fmt.Sprintf("\n\tmock %q", mock.name)

		return errors.New(errorMessage)
	}

	mock.methods = make([]methodInfo, 0, mock.typ.NumMethods())
	for i := 0; i < mock.typ.NumMethods(); i++ {
		fun := mock.typ.Method(i)
		methodInfo := methodInfo{
			typ:       fun,
			fieldName: fieldNameFormatter(fun.Name()),
		}

		sig := fun.Type().(*types.Signature)

		methodInfo.params = make([]paramInfo, 0, sig.Params().Len())
		for i := 0; i < sig.Params().Len(); i++ {
			param := sig.Params().At(i)

			methodInfo.params = append(methodInfo.params, paramInfo{
				typ:        param,
				isVariadic: i+1 == sig.Params().Len() && sig.Variadic(),
			})
		}

		methodInfo.results = make([]resultInfo, 0, sig.Results().Len())
		for i := 0; i < sig.Results().Len(); i++ {
			result := sig.Results().At(i)

			methodInfo.results = append(methodInfo.results, resultInfo{
				typ: result,
			})
		}

		mock.methods = append(mock.methods, methodInfo)
	}

	g.mocks = append(g.mocks, mock)

	return nil
}

func complete(inter *types.Interface) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("%v", rec)
		}
	}()

	inter.Complete()

	return nil
}

func typeCode(stmt *jen.Statement, t types.Type) jen.Code {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Name() {
		case "bool":
			return stmt.Bool()
		case "int":
			return stmt.Int()
		case "int8":
			return stmt.Int8()
		case "int16":
			return stmt.Int16()
		case "int32":
			return stmt.Int32()
		case "int64":
			return stmt.Int64()
		case "uint":
			return stmt.Uint()
		case "uint8":
			return stmt.Uint8()
		case "uint16":
			return stmt.Uint16()
		case "uint32":
			return stmt.Uint32()
		case "uint64":
			return stmt.Uint64()
		case "uintptr":
			return stmt.Uintptr()
		case "float32":
			return stmt.Float32()
		case "float64":
			return stmt.Float64()
		case "complex64":
			return stmt.Complex64()
		case "complex128":
			return stmt.Complex128()
		case "string":
			return stmt.String()
		case "Pointer":
			return stmt.Qual("unsafe", "Pointer")
		case "byte":
			return stmt.Byte()
		case "rune":
			return stmt.Rune()
		}
	case *types.Array:
		return typeCode(stmt.Index(jen.Lit(t.Len())), t.Elem())
	case *types.Slice:
		return typeCode(stmt.Index(), t.Elem())
	case *types.Struct:
		return stmt.StructFunc(func(g *jen.Group) {
			for i := 0; i < t.NumFields(); i++ {
				f := t.Field(i)
				g.Do(func(s *jen.Statement) {
					if f.Anonymous() {
						typeCode(s, f.Type())
					} else {
						typeCode(s.Id(f.Name()), f.Type())
					}
				})
			}
		})
	case *types.Pointer:
		return typeCode(stmt.Op("*"), t.Elem())
	case *types.Tuple:
		return stmt.ValuesFunc(func(g *jen.Group) {
			typeTupleCode(g, t, false)
		})
	case *types.Signature:
		return stmt.Func().ParamsFunc(func(g *jen.Group) {
			typeTupleCode(g, t.Params(), t.Variadic())
		}).ParamsFunc(func(g *jen.Group) {
			typeTupleCode(g, t.Results(), false)
		})
	case *types.Interface:
		return stmt.InterfaceFunc(func(g *jen.Group) {
			for i := 0; i < t.NumEmbeddeds(); i++ {
				e := t.EmbeddedType(i)
				g.Do(func(s *jen.Statement) {
					typeCode(s, e)
				})
			}
			for i := 0; i < t.NumExplicitMethods(); i++ {
				m := t.ExplicitMethod(i)
				sig := m.Type().(*types.Signature)

				g.Do(func(s *jen.Statement) {
					s.Id(m.Name()).ParamsFunc(func(g *jen.Group) {
						typeTupleCode(g, sig.Params(), sig.Variadic())
					}).ParamsFunc(func(g *jen.Group) {
						typeTupleCode(g, sig.Results(), false)
					})
				})
			}
		})
	case *types.Map:
		return typeCode(stmt.Map(typeCode(nil, t.Key())), t.Elem())
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			return typeCode(stmt.Chan(), t.Elem())
		case types.RecvOnly:
			return typeCode(stmt.Op("<-").Chan(), t.Elem())
		default:
			return typeCode(stmt.Chan().Op("<-"), t.Elem())
		}
	case *types.Named:
		if i := strings.LastIndex(t.String(), "."); i >= 0 {
			return stmt.Qual(t.String()[:i], t.String()[i+1:])
		}
		return stmt.Id(t.String())
	}
	return stmt
}

func typeTupleCode(g *jen.Group, t *types.Tuple, variadic bool) {
	for i := 0; i < t.Len(); i++ {
		g.Do(func(s *jen.Statement) {
			v := t.At(i)
			if variadic && i+1 == t.Len() {
				typeCode(s.Id("").Op("..."), v.Type().(*types.Slice).Elem())
			} else {
				typeCode(s.Id(""), v.Type())
			}
		})
	}
}

func newFieldNameFormatter(prefix, suffix string) func(string) string {
	return func(field string) string {
		return prefix + field + suffix
	}
}
