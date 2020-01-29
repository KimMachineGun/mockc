package mockc

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/tools/go/packages"
)

const (
	mockcPath = "github.com/KimMachineGun/mockc"
)

type Mockc struct {
	wd       string
	patterns []string
}

func New(wd string, patterns []string) *Mockc {
	return &Mockc{
		wd:       wd,
		patterns: patterns,
	}
}

func (m *Mockc) Execute(ctx context.Context) error {
	pkgs, err := m.loadPackages(ctx, m.wd, m.patterns)
	if err != nil {
		return fmt.Errorf("cannot load package: %v", err)
	}

	for _, pkg := range pkgs {
		if _, ok := pkg.Imports[mockcPath]; !ok {
			continue
		}

		g, err := newGenerator(pkg)
		if err != nil {
			return fmt.Errorf("package \"%s\": cannot create generator: %v", pkg.PkgPath, err)
		}

		err = g.generate()
		if err != nil {
			return fmt.Errorf("package \"%s\": cannot generate mocks: %v", pkg.PkgPath, err)
		}
	}

	return nil
}

func (m *Mockc) loadPackages(ctx context.Context, wd string, patterns []string) ([]*packages.Package, error) {
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
		Tests:      true,
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
