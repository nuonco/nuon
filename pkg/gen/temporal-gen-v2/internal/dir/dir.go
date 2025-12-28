package dir

import (
	"context"
	"fmt"

	"golang.org/x/tools/go/packages"
)

// Package represents a loaded Go package with its syntax trees
type Package struct {
	Pkg *packages.Package
}

// LoadPackages loads the packages in the given directory
func LoadPackages(ctx context.Context, dir string) ([]*Package, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Tests:   false,
	}

	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	var result []*Package
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("package has errors: %v", pkg.Errors)
		}
		result = append(result, &Package{Pkg: pkg})
	}

	return result, nil
}

// LoadPackage loads a single package by its path
func LoadPackage(ctx context.Context, path string) (*Package, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Tests:   false,
	}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load package %s: %w", path, err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package %s not found", path)
	}

	// With a specific package path, we should get exactly one package unless there are errors
	// If we get multiple (e.g. test variants), we take the first one that matches the ID if possible,
	// or just the first one since we disabled Tests.
	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		return nil, fmt.Errorf("package %s has errors: %v", path, pkg.Errors)
	}

	return &Package{Pkg: pkg}, nil
}

