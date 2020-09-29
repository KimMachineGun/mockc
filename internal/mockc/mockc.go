package mockc

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	mockcPath              = "github.com/KimMachineGun/mockc"
	defaultDestination     = "mockc_gen.go"
	defaultFieldNamePrefix = "_"
	defaultFieldNameSuffix = ""
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

		generators, err := newParser(pkg).Parse()
		if err != nil {
			return err
		}

		for _, generator := range generators {
			err = generator.Generate("mockc")
			if err != nil {
				return fmt.Errorf("package %q: cannot generate mock: %v", pkg.PkgPath, err)
			}
		}
	}

	return nil
}

func GenerateWithFlags(ctx context.Context, wd string, destination string, name string, withConstructor bool, fieldNamePrefix string, fieldNameSuffix string, interfacePatterns []string) error {
	destination, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("cannot convert destination into absolute path: %v", err)
	}

	destinationDir, fileName := filepath.Split(destination)
	if filepath.Ext(fileName) != ".go" {
		return fmt.Errorf("destination should be a go file: %s", fileName)
	} else if destinationDir == "" {
		destinationDir = "."
	}

	pkgs, err := loadPackages(ctx, wd, []string{destinationDir})
	if err != nil {
		return fmt.Errorf("cannot load destination package: %v", err)
	} else if len(pkgs) != 1 {
		return fmt.Errorf("muptile destination packages are loaded: %v", pkgs)
	}

	generator := newGenerator(pkgs[0], destination)

	err = generator.addMockWithFlags(ctx, wd, name, withConstructor, fieldNamePrefix, fieldNameSuffix, interfacePatterns)
	if err != nil {
		return err
	}

	err = generator.Generate(fmt.Sprintf("mockc \"-destination=%s\" \"-name=%s\" \"-withConstructor=%t\" \"-fieldNamePrefix=%s\" \"-fieldNameSuffix=%s\" \"%s\"", fileName, name, withConstructor, fieldNamePrefix, fieldNameSuffix, strings.Join(interfacePatterns, " ")))
	if err != nil {
		return fmt.Errorf("cannot generate mock: %v", err)
	}

	return nil
}
