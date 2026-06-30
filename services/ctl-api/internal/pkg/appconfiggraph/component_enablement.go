// Package appconfiggraph builds the component dependency graph of an install's
// pinned app config and resolves effective-enabled state and cascade ordering
// over it. It is pure graph logic: given the config-connection snapshot and the
// install's enabled inputs, it answers which components should be deployed and
// in what order, with no database, FX, or queue dependencies. The install
// workflow step generator and the service-side toggle validation share it.
//
// The package is split into two layers: Graph holds the pure dependency
// structure (see graph.go), and ComponentEnablementResolver overlays the
// install's per-component toggles on top of a Graph to resolve effective-enabled
// state.
package appconfiggraph

import (
	"sort"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ComponentEnablementResolver overlays an install's per-component toggles on a
// Graph to resolve the effective-enabled state of each component. A component is
// effectively enabled only when its own toggle is on (non-toggleable components
// are always own-enabled) AND every component it depends on is effectively
// enabled:
//
//	effectiveEnabled(C) = ownEnabled(C) && all(effectiveEnabled(dep) for dep in deps(C))
//
// Own-enabled state is sourced from the install's inputs via the reserved
// synthetic enabled input (see ComponentEnabledFromInputs), which is the single
// source of truth for a component's toggle. The embedded *Graph supplies the
// structural queries (DepEdges, DirectDependents, TopoSort, etc.).
type ComponentEnablementResolver struct {
	*Graph
	cccByID       map[string]*app.ComponentConfigConnection
	enabledInputs map[string]*string
	cache         map[string]bool
}

// NewComponentEnablementResolver builds a resolver from a component-ID keyed set
// of config connections (the install's pinned app config snapshot) and the
// install's latest input values, which carry the per-component enabled toggles.
func NewComponentEnablementResolver(cccByID map[string]*app.ComponentConfigConnection, enabledInputs map[string]*string) *ComponentEnablementResolver {
	return &ComponentEnablementResolver{
		Graph:         NewGraph(cccByID),
		cccByID:       cccByID,
		enabledInputs: enabledInputs,
		cache:         make(map[string]bool),
	}
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

	if !app.ComponentEnabledFromInputs(r.enabledInputs, r.cccByID[compID]) {
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

// OwnEnabled reports whether the component's own toggle is on, ignoring its
// dependency closure. Non-toggleable components are always own-enabled.
func (r *ComponentEnablementResolver) OwnEnabled(compID string) bool {
	return app.ComponentEnabledFromInputs(r.enabledInputs, r.cccByID[compID])
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
