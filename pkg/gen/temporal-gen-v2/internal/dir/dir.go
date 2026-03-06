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
	// Use AST-only parsing by default for speed and to avoid compilation errors
	// We extract all the type information we need directly from the AST
	cfg := &packages.Config{
		Context: ctx,
		// Only load syntax and basic info, skip type checking
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load package %s: %w", path, err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package %s not found", path)
	}

	return &Package{Pkg: pkgs[0]}, nil
}
