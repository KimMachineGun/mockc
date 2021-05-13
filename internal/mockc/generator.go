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
	iface, err := overlapInterfaces(interfaces)
	if err != nil {
		errorMessage := err.Error()
		errorMessage += fmt.Sprintf("\n\tmock %q", name)

		return errors.New(errorMessage)
	}

	methods := make([]methodInfo, iface.NumMethods())
	for i := 0; i < iface.NumMethods(); i++ {
		method := iface.Method(i)
		sig := method.Type().(*types.Signature)

		params := make([]paramInfo, sig.Params().Len())
		for i := 0; i < sig.Params().Len(); i++ {
			params[i] = paramInfo{
				typ:        sig.Params().At(i),
				isVariadic: i+1 == sig.Params().Len() && sig.Variadic(),
			}
		}

		results := make([]resultInfo, sig.Results().Len())
		for i := 0; i < sig.Results().Len(); i++ {
			results[i] = resultInfo{
				typ: sig.Results().At(i),
			}
		}

		methods[i] = methodInfo{
			typ:       method,
			fieldName: fieldNameFormatter(method.Name()),
			params:    params,
			results:   results,
		}
	}

	g.mocks = append(g.mocks, mockInfo{
		typ:         iface,
		name:        name,
		constructor: constructor,
		methods:     methods,
	})

	return nil
}

func overlapInterfaces(interfaces []types.Type) (iface *types.Interface, err error) {
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

	iface = types.NewInterfaceType(methods, embeddeds)
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("%v", rec)
		}
	}()
	iface.Complete()

	return iface, err
}

func newFieldNameFormatter(prefix, suffix string) func(string) string {
	return func(field string) string {
		return prefix + field + suffix
	}
}
