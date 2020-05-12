package mockc

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func validateInterface(inter *types.Interface, isExternalInterface bool) error {
	for i := 0; i < inter.NumMethods(); i++ {
		method := inter.Method(i)
		if isExternalInterface && !method.Exported() {
			return fmt.Errorf("cannot implement non-exported method: %s", method.FullName())
		}
	}

	return nil
}

type interfaceFinder struct {
	pkg     *packages.Package
	targets []string
	result  map[string]*types.Interface
}

func newInterfaceFinder(pkg *packages.Package, targets []string) *interfaceFinder {
	return &interfaceFinder{
		pkg:     pkg,
		targets: targets,
		result:  map[string]*types.Interface{},
	}
}

func (f *interfaceFinder) Visit(node ast.Node) ast.Visitor {
	n, ok := node.(*ast.TypeSpec)
	if !ok {
		return f
	}

	inter, ok := f.pkg.TypesInfo.TypeOf(n.Type).(*types.Interface)
	if !ok {
		return f
	}

	for _, interfaceName := range f.targets {
		if interfaceName == n.Name.Name {
			f.result[interfaceName] = inter
		}
	}

	return f
}
