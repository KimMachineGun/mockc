package mockc

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func loadPackages(ctx context.Context, wd string, patterns []string) ([]*packages.Package, error) {
	if len(patterns) == 0 {
		patterns = []string{"."}
	}

	patterns = append(make([]string, 0, len(patterns)), patterns...)
	for i, pattern := range patterns {
		patterns[i] = "pattern=" + pattern
	}

	cfg := &packages.Config{
		Context:    ctx,
		Mode:       packages.LoadAllSyntax,
		Dir:        wd,
		BuildFlags: []string{"-tags=mockc"},
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}

	var errs []error
	for _, p := range pkgs {
		for _, e := range p.Errors {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		var errMessage string
		for _, err := range errs {
			errMessage += fmt.Sprintf("\n\t%v", err)
		}

		return nil, errors.New(errMessage)
	}

	return pkgs, nil
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
