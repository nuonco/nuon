package app

import (
	"sort"

	"github.com/nuonco/nuon/pkg/config/refs"
)

// ComponentEnablementResolver resolves the effective-enabled state of the
// components in an install's pinned app config, given the install's enabled
// inputs. A component is effectively enabled only when its own toggle is on
// (non-toggleable components are always own-enabled) AND every component it
// depends on is effectively enabled:
//
//	effectiveEnabled(C) = ownEnabled(C) && all(effectiveEnabled(dep) for dep in deps(C))
//
// Dependencies are the UNION of declared dependencies (ComponentDependencyIDs)
// and components whose outputs C references (Refs of type component), scoped to
// the components present in the provided set.
type ComponentEnablementResolver struct {
	cccByID       map[string]*ComponentConfigConnection
	enabledInputs map[string]*string
	depEdges      map[string]map[string]struct{}
	cache         map[string]bool
}

// NewComponentEnablementResolver builds a resolver from a component-ID keyed set
// of config connections (the install's pinned app config snapshot) and the
// install's latest enabled inputs.
func NewComponentEnablementResolver(cccByID map[string]*ComponentConfigConnection, enabledInputs map[string]*string) *ComponentEnablementResolver {
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

	return &ComponentEnablementResolver{
		cccByID:       cccByID,
		enabledInputs: enabledInputs,
		depEdges:      edges,
		cache:         make(map[string]bool),
	}
}

// DepEdges returns the union dependency set (deps + output refs) per component.
func (r *ComponentEnablementResolver) DepEdges() map[string]map[string]struct{} {
	return r.depEdges
}

// EffectiveEnabled reports whether the component should be deployed given its
// own toggle and its dependency closure. Results are memoized; the visiting set
// guards against cycles defensively (the app config graph is a DAG).
func (r *ComponentEnablementResolver) EffectiveEnabled(compID string) bool {
	return r.compute(compID, make(map[string]struct{}))
}

func (r *ComponentEnablementResolver) compute(compID string, visiting map[string]struct{}) bool {
	if v, ok := r.cache[compID]; ok {
		return v
	}
	// Fail closed on a dependency cycle. The app config graph is a DAG, so this
	// is defensive; returning false (rather than true) avoids caching a
	// transitively-incorrect "enabled" for a node still mid-traversal.
	if _, cycle := visiting[compID]; cycle {
		return false
	}

	if !ComponentEnabledFromInputs(r.enabledInputs, r.cccByID[compID]) {
		r.cache[compID] = false
		return false
	}

	visiting[compID] = struct{}{}
	defer delete(visiting, compID)

	res := true
	for dep := range r.depEdges[compID] {
		if !r.compute(dep, visiting) {
			res = false
			break
		}
	}

	r.cache[compID] = res
	return res
}

// DisabledDependencies returns the direct dependencies of compID that are not
// effectively enabled. Used to build user-facing "enable X first" errors.
func (r *ComponentEnablementResolver) DisabledDependencies(compID string) []string {
	var disabled []string
	for dep := range r.depEdges[compID] {
		if !r.EffectiveEnabled(dep) {
			disabled = append(disabled, dep)
		}
	}
	sort.Strings(disabled)
	return disabled
}

// TransitiveDependentsClosure returns the given roots plus every component that
// transitively depends on any root (the blast radius of toggling those roots).
func (r *ComponentEnablementResolver) TransitiveDependentsClosure(rootIDs []string) []string {
	dependents := make(map[string][]string, len(r.depEdges))
	for compID, deps := range r.depEdges {
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
func (r *ComponentEnablementResolver) TopoSort(ids []string) []string {
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

		deps := make([]string, 0, len(r.depEdges[id]))
		for dep := range r.depEdges[id] {
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
func (r *ComponentEnablementResolver) ReverseTopoSort(ids []string) []string {
	order := r.TopoSort(ids)
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}
	return order
}
