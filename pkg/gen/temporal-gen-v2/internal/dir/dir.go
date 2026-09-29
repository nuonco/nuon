package dir

import (
	"context"
	"fmt"
	"sort"

	"golang.org/x/tools/go/packages"
)

type Package struct {
	Pkg *packages.Package
}

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

func LoadPackage(ctx context.Context, path string) (*Package, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax,
		Tests:   false,
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

func LoadPackageLevels(ctx context.Context, pattern string) ([][]*Package, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedImports,
		Tests:   false,
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
