package dir

import (
	"context"
	"fmt"
	"sort"

	"golang.org/x/tools/go/packages"
)

// buildDependencyGraph builds a directed dependency graph from a set of packages.
// graph[dep] = []dependents means dep must be processed before each dependent.
// inDegree[pkg] = number of target-set dependencies pkg has.
func buildDependencyGraph(pkgs []*packages.Package, targetPkgs map[string]bool) (graph map[string][]string, inDegree map[string]int) {
	graph = make(map[string][]string)
	inDegree = make(map[string]int)

	for id := range targetPkgs {
		inDegree[id] = 0
	}

	for _, pkg := range pkgs {
		for _, imp := range pkg.Imports {
			if targetPkgs[imp.ID] {
				graph[imp.ID] = append(graph[imp.ID], pkg.ID)
				inDegree[pkg.ID]++
			}
		}
	}

	return graph, inDegree
}

// GetDependencyOrder returns the package paths in topological order (dependencies first).
// It only includes packages within the specified directory pattern.
func GetDependencyOrder(ctx context.Context, dir string) ([]string, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedName | packages.NeedImports,
		Tests:   false,
	}

	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	pkgMap := make(map[string]*packages.Package)
	targetPkgs := make(map[string]bool)
	for _, pkg := range pkgs {
		pkgMap[pkg.ID] = pkg
		targetPkgs[pkg.ID] = true
	}

	graph, inDegree := buildDependencyGraph(pkgs, targetPkgs)

	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	var result []string
	processedCount := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		processedCount++

		if p, ok := pkgMap[current]; ok {
			result = append(result, p.PkgPath)
		}

		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
		sort.Strings(queue)
	}

	if processedCount != len(targetPkgs) {
		return nil, fmt.Errorf("cycle detected in package dependencies: processed %d of %d packages", processedCount, len(targetPkgs))
	}

	return result, nil
}

// GetDependencyLevels returns packages grouped by processing level.
// All packages within the same level have no dependencies on each other
// and can be processed concurrently. Levels must be processed in order.
func GetDependencyLevels(ctx context.Context, dir string) ([][]string, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedName | packages.NeedImports,
		Tests:   false,
	}

	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	pkgMap := make(map[string]*packages.Package)
	targetPkgs := make(map[string]bool)
	for _, pkg := range pkgs {
		pkgMap[pkg.ID] = pkg
		targetPkgs[pkg.ID] = true
	}

	graph, inDegree := buildDependencyGraph(pkgs, targetPkgs)

	// Collect initial zero-degree nodes as level 0
	var currentLevel []string
	for id, degree := range inDegree {
		if degree == 0 {
			currentLevel = append(currentLevel, id)
		}
	}
	sort.Strings(currentLevel)

	var levels [][]string
	processedCount := 0

	for len(currentLevel) > 0 {
		// Resolve IDs to PkgPaths for this level
		var levelPaths []string
		for _, id := range currentLevel {
			if p, ok := pkgMap[id]; ok {
				levelPaths = append(levelPaths, p.PkgPath)
			}
		}
		levels = append(levels, levelPaths)
		processedCount += len(currentLevel)

		// Compute next level by processing all nodes in the current level
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
