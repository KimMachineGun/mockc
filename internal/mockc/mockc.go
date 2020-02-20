package mockc

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	mockcPath = "github.com/KimMachineGun/mockc"
)

func Generate(ctx context.Context, wd string, patterns []string) error {
	pkgs, err := loadPackages(ctx, wd, patterns)
	if err != nil {
		return fmt.Errorf("cannot load packages: %v", err)
	}

	for _, pkg := range pkgs {
		if _, ok := pkg.Imports[mockcPath]; !ok {
			continue
		}

		g := newGenerator(pkg, filepath.Join(filepath.Dir(pkg.GoFiles[0]), "mockc_gen.go"))

		err = g.loadMocks()
		if err != nil {
			return fmt.Errorf("package %q: cannot load mocks: %v", pkg.PkgPath, err)
		}

		err = g.generate("mockc")
		if err != nil {
			return fmt.Errorf("package %q: cannot generate mocks: %v", pkg.PkgPath, err)
		}
	}

	return nil
}

func GenerateWithFlags(ctx context.Context, wd string, name string, destination string, fieldNamePrefix string, fieldNameSuffix string, interfacePatterns []string) error {
	destination, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("cannot convert destination into absolute path: %v", err)
	}

	destinationDir, fileName := filepath.Split(destination)
	if filepath.Ext(fileName) != ".go" {
		return fmt.Errorf("destination file should be go file: %s", fileName)
	} else if destinationDir == "" {
		destinationDir = "."
	}

	pkgs, err := loadPackages(ctx, wd, []string{destinationDir})
	if err != nil {
		return fmt.Errorf("cannot load destination package: %v", err)
	} else if len(pkgs) != 1 {
		return fmt.Errorf("muptile destination packages are loaded: %v", pkgs)
	}

	g := newGenerator(pkgs[0], destination)

	err = g.loadMockWithFlags(ctx, wd, name, fieldNamePrefix, fieldNameSuffix, interfacePatterns)
	if err != nil {
		return fmt.Errorf("cannot load mock: %v", err)
	}

	err = g.generate(fmt.Sprintf("mockc -name=%s -destination=%s %s", name, destination, strings.Join(interfacePatterns, " ")))
	if err != nil {
		return fmt.Errorf("cannot generate mock: %v", err)
	}

	return nil
}
