package dir

import (
	"context"
	"fmt"
	"sort"

	"golang.org/x/tools/go/packages"
)

// GetDependencyOrder returns the package paths in topological order (dependencies first).
// It only includes packages within the specified directory pattern.
func GetDependencyOrder(ctx context.Context, dir string) ([]string, error) {
	// Load with minimal information needed for dependency graph
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedName | packages.NeedImports,
		Tests:   false,
	}

	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	// Map package ID to package
	pkgMap := make(map[string]*packages.Package)
	// Set of package IDs that are part of our target set
	targetPkgs := make(map[string]bool)

	for _, pkg := range pkgs {
		pkgMap[pkg.ID] = pkg
		targetPkgs[pkg.ID] = true
	}

	// Build dependency graph
	// graph[dependency] = []dependents
	// If A imports B, then B is a dependency of A.
	// We want B to be processed before A.
	// So we draw an edge B -> A.
	// In-degree of A increases.
	graph := make(map[string][]string)
	// inDegree[pkgID] = count of dependencies within target set
	inDegree := make(map[string]int)

	// Initialize inDegree for all target packages
	for id := range targetPkgs {
		inDegree[id] = 0
	}

	for _, pkg := range pkgs {
		for _, imp := range pkg.Imports {
			// We only care if the dependency is also in our target set
			if targetPkgs[imp.ID] {
				graph[imp.ID] = append(graph[imp.ID], pkg.ID)
				inDegree[pkg.ID]++
			}
		}
	}

	// Kahn's algorithm
	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	// Sort initial queue to ensure deterministic output
	sort.Strings(queue)

	var result []string
	processedCount := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		processedCount++

		// Add to result
		if p, ok := pkgMap[current]; ok {
			result = append(result, p.PkgPath)
		}

		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
		// Sort queue again for determinism
		sort.Strings(queue)
	}

	if processedCount != len(targetPkgs) {
		return nil, fmt.Errorf("cycle detected in package dependencies: processed %d of %d packages", processedCount, len(targetPkgs))
	}

	return result, nil
}
