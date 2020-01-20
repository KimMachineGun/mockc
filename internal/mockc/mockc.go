package mockc

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/tools/go/packages"
)

const (
	packageName = "github.com/KimMachineGun/mockc"
)

type Mockc struct {
	wd       string
	patterns []string

	pkgs []Package
}

func NewMockc(wd string, patterns []string) *Mockc {
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

	mockPackages := make([]*Package, 0, len(pkgs))

	for _, pkg := range pkgs {
		p, err := newPackage(pkg)
		if err != nil {
			return fmt.Errorf("cannot create Package: %v", err)
		}

		if p.hasMockc() {
			mockPackages = append(mockPackages, p)
		}
	}

	for _, p := range mockPackages {
		if err := p.GenerateResult(); err != nil {
			return fmt.Errorf("cannot generate results: %v", err)
		}
	}

	return nil
}

func (m *Mockc) loadPackages(ctx context.Context, wd string, patterns []string) ([]*packages.Package, error) {
	if len(patterns) == 0 {
		patterns = []string{"."}
	}

	cfg := &packages.Config{
		Context:    ctx,
		Mode:       packages.LoadAllSyntax,
		Dir:        wd,
		Tests:      true,
		BuildFlags: []string{"-tags=mockc"},
	}

	escaped := make([]string, len(patterns))
	for i := range patterns {
		escaped[i] = "pattern=" + patterns[i]
	}

	pkgs, err := packages.Load(cfg, escaped...)
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
