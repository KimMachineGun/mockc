package mockc

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func validateInterface(pkg *packages.Package, inter *types.Interface, isExternalInterface bool) error {
	for i := 0; i < inter.NumMethods(); i++ {
		method := inter.Method(i)
		if isExternalInterface && !method.Exported() {
			return fmt.Errorf("cannot implement external interface with unexported method: %s", method.FullName())
		}

		sig := method.Type().(*types.Signature)
		for j := 0; j < sig.Params().Len(); j++ {
			p := sig.Params().At(j)

			if !isAccessible(pkg, p) {
				return fmt.Errorf("cannot implement interface using unexported type: %s %s", method.FullName(), p.Type())
			}
		}

		for j := 0; j < sig.Results().Len(); j++ {
			r := sig.Results().At(j)

			if !isAccessible(pkg, r) {
				return fmt.Errorf("cannot implement interface using unexported type: %s %s", method.FullName(), r.Type())
			}
		}
	}

	return nil
}

func isAccessible(pkg *packages.Package, obj types.Object) bool {
	return pkg.Types.Path() == obj.Pkg().Path() || obj.Exported()
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
