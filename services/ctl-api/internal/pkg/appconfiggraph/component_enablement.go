package appconfiggraph

import (
	"sort"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ComponentEnablementResolver struct {
	*Graph
	cccByID       map[string]*app.ComponentConfigConnection
	enabledInputs map[string]*string
	cache         map[string]bool
}

func NewComponentEnablementResolver(cccByID map[string]*app.ComponentConfigConnection, enabledInputs map[string]*string) *ComponentEnablementResolver {
	return &ComponentEnablementResolver{
		Graph:         NewGraph(cccByID),
		cccByID:       cccByID,
		enabledInputs: enabledInputs,
		cache:         make(map[string]bool),
	}
}

func (r *ComponentEnablementResolver) EffectiveEnabled(compID string) bool {
	return r.compute(compID, make(map[string]struct{}))
}

func (r *ComponentEnablementResolver) compute(compID string, visiting map[string]struct{}) bool {
	if v, ok := r.cache[compID]; ok {
		return v
	}
	// why: Fail closed on a dependency cycle. The app config graph is a DAG, so this
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

func (r *ComponentEnablementResolver) OwnEnabled(compID string) bool {
	return app.ComponentEnabledFromInputs(r.enabledInputs, r.cccByID[compID])
}

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
