package appconfiggraph

import (
	"sort"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Graph is the dependency graph of the components in an install's pinned app
// config. An edge C -> D means C depends on D, where dependencies are the UNION
// of declared dependencies (ComponentDependencyIDs) and components whose outputs
// C references (Refs of type component), scoped to the components present in the
// provided set. It is pure structure: it carries no toggle/enabled state and has
// no database, FX, or queue dependencies. Effective-enabled resolution is layered
// on top by ComponentEnablementResolver.
type Graph struct {
	depEdges map[string]map[string]struct{}
}

// NewGraph builds the dependency graph from a component-ID keyed set of config
// connections (the install's pinned app config snapshot).
func NewGraph(cccByID map[string]*app.ComponentConfigConnection) *Graph {
	nameToID := make(map[string]string, len(cccByID))
	for id, ccc := range cccByID {
		if ccc == nil {
			continue
		}
		nameToID[ccc.Component.Name] = id
	}

	edges := make(map[string]map[string]struct{}, len(cccByID))
	for id, ccc := range cccByID {
		if ccc == nil {
			continue
		}
		set := make(map[string]struct{})
		for _, dep := range ccc.ComponentDependencyIDs {
			if _, ok := cccByID[dep]; ok {
				set[dep] = struct{}{}
			}
		}
		for _, r := range ccc.Refs {
			if r.Type != refs.RefTypeComponents {
				continue
			}
			depID, ok := nameToID[r.Name]
			if !ok {
				continue
			}
			if _, ok := cccByID[depID]; ok {
				set[depID] = struct{}{}
			}
		}
		edges[id] = set
	}

	return &Graph{depEdges: edges}
}

// DepEdges returns the union dependency set (deps + output refs) per component.
func (g *Graph) DepEdges() map[string]map[string]struct{} {
	return g.depEdges
}

// DirectDependents returns the components that directly declare compID as a
// dependency (the reverse of DepEdges for a single node).
func (g *Graph) DirectDependents(compID string) []string {
	var dependents []string
	for id, deps := range g.depEdges {
		if _, ok := deps[compID]; ok {
			dependents = append(dependents, id)
		}
	}
	sort.Strings(dependents)
	return dependents
}

// TransitiveDependentsClosure returns the given roots plus every component that
// transitively depends on any root (the blast radius of toggling those roots).
func (g *Graph) TransitiveDependentsClosure(rootIDs []string) []string {
	dependents := make(map[string][]string, len(g.depEdges))
	for compID, deps := range g.depEdges {
		for dep := range deps {
			dependents[dep] = append(dependents[dep], compID)
		}
	}
	// Sort each adjacency list so BFS yields a deterministic order — this slice
	// feeds a Temporal activity's arguments, so a stable order is required to
	// avoid replay nondeterminism.
	for dep := range dependents {
		sort.Strings(dependents[dep])
	}

	seen := make(map[string]struct{})
	order := make([]string, 0)
	queue := append([]string(nil), rootIDs...)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		order = append(order, id)
		queue = append(queue, dependents[id]...)
	}
	return order
}

// TopoSort orders ids so dependencies come before the components that depend on
// them (deploy order). Ties (components with no in-set dependency between them)
// preserve the caller's input order, so unrelated components are not reordered.
func (g *Graph) TopoSort(ids []string) []string {
	inSet := make(map[string]struct{}, len(ids))
	idx := make(map[string]int, len(ids))
	for i, id := range ids {
		inSet[id] = struct{}{}
		if _, ok := idx[id]; !ok {
			idx[id] = i
		}
	}
	byInput := func(a, b string) bool { return idx[a] < idx[b] }

	visited := make(map[string]struct{}, len(ids))
	order := make([]string, 0, len(ids))

	var visit func(string)
	visit = func(id string) {
		if _, ok := visited[id]; ok {
			return
		}
		visited[id] = struct{}{}

		deps := make([]string, 0, len(g.depEdges[id]))
		for dep := range g.depEdges[id] {
			if _, ok := inSet[dep]; ok {
				deps = append(deps, dep)
			}
		}
		sort.SliceStable(deps, func(i, j int) bool { return byInput(deps[i], deps[j]) })
		for _, dep := range deps {
			visit(dep)
		}
		order = append(order, id)
	}

	roots := append([]string(nil), ids...)
	sort.SliceStable(roots, func(i, j int) bool { return byInput(roots[i], roots[j]) })
	for _, id := range roots {
		visit(id)
	}
	return order
}

// ReverseTopoSort orders ids so dependents come before the dependencies they
// rely on (teardown order).
func (g *Graph) ReverseTopoSort(ids []string) []string {
	order := g.TopoSort(ids)
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}
	return order
}
