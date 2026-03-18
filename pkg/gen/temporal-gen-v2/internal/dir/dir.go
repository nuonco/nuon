package dir

import (
	"context"
	"fmt"
	"sort"

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

// LoadPackageLevels loads all packages matching the pattern in a single packages.Load
// call and returns them grouped by dependency level. Packages within the same level
// have no intra-level dependencies and can be processed concurrently.
//
// This is significantly faster than calling LoadPackage per package because it
// uses a single Go toolchain invocation instead of one per package.
func LoadPackageLevels(ctx context.Context, pattern string) ([][]*Package, error) {
	cfg := &packages.Config{
		Context: ctx,
		// Load syntax + imports in one shot: enough for both dependency ordering
		// and annotation-based code generation (AST-only, no type checking).
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedImports,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	pkgMap := make(map[string]*packages.Package, len(pkgs))
	targetPkgs := make(map[string]bool, len(pkgs))
	for _, pkg := range pkgs {
		pkgMap[pkg.ID] = pkg
		targetPkgs[pkg.ID] = true
	}

	graph, inDegree := buildDependencyGraph(pkgs, targetPkgs)

	var currentLevel []string
	for id, degree := range inDegree {
		if degree == 0 {
			currentLevel = append(currentLevel, id)
		}
	}
	sort.Strings(currentLevel)

	var levels [][]*Package
	processedCount := 0

	for len(currentLevel) > 0 {
		var levelPkgs []*Package
		for _, id := range currentLevel {
			if p, ok := pkgMap[id]; ok {
				levelPkgs = append(levelPkgs, &Package{Pkg: p})
			}
		}
		levels = append(levels, levelPkgs)
		processedCount += len(currentLevel)

		var nextLevel []string
		for _, id := range currentLevel {
			for _, dependent := range graph[id] {
				inDegree[dependent]--
				if inDegree[dependent] == 0 {
					nextLevel = append(nextLevel, dependent)
				}
			}
		}
		sort.Strings(nextLevel)
		currentLevel = nextLevel
	}

	if processedCount != len(targetPkgs) {
		return nil, fmt.Errorf("cycle detected in package dependencies: processed %d of %d packages", processedCount, len(targetPkgs))
	}

	return levels, nil
}
