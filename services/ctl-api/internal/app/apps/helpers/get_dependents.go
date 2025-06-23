package helpers

import (
	"context"
	"fmt"
	"slices"

	"github.com/dominikbraun/graph"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) GetDependentComponents(ctx context.Context, appID, compRootID string) ([]app.Component, error) {
	g, cmps, err := h.GetDependencyGraph(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get graph: %w", err)
	}
	depCmps, err := h.GetDependentComponentsByGraph(compRootID, g, cmps)
	if err != nil {
		return nil, fmt.Errorf("unable to get dependent components: %w", err)
	}
	return depCmps, nil
}

func (h *Helpers) GetComponentsDependents(compRootID string, cmps []app.Component) []app.Component {
	var depCmps []app.Component
	for _, cmp := range cmps {
		for _, dep := range cmp.Dependencies {
			if dep.ID == compRootID {
				depCmps = append(depCmps, cmp)
			}
		}
	}
	return depCmps
}

func (h *Helpers) GetDependentComponentsByGraph(compRootID string, g graph.Graph[string, *app.Component], cmps []app.Component) ([]app.Component, error) {
	cmpsById := make(map[string]app.Component)
	for _, c := range cmps {
		cmpsById[c.ID] = c
	}

	depsCmpIds := make([]string, 0)
	if err := graph.BFS(g, compRootID, func(compID string) bool {
		if compID == compRootID {
			return false
		}

		if !slices.Contains(depsCmpIds, compID) {
			depsCmpIds = append(depsCmpIds, compID)
		}

		return false
	}); err != nil {
		return nil, fmt.Errorf("unable to build app graph: %w", err)
	}

	var err error
	depCmps := make([]app.Component, 0)
	for _, id := range depsCmpIds {
		comp, ok := cmpsById[id]
		if !ok {
			return nil, fmt.Errorf("unable to get component: %w", err)
		}
		depCmps = append(depCmps, comp)
	}

	slices.Reverse(depCmps)
	return depCmps, nil
}

func (h *Helpers) GetInvertedDependentComponents(ctx context.Context, appID, compRootID string) ([]app.Component, error) {
	g, cmps, err := h.GetInvertedDependencyGraph(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get graph: %w", err)
	}

	return h.getInvertedDependentComponentsFromComponents(ctx, &g, &cmps, compRootID)

}

// GetInvertedDependentByComponentConfigVersion retrieves the components that depend on the given component root ID
func (h *Helpers) GetInvertedDependentByComponentConfigVersion(ctx context.Context, appID string, configVersion int, compRootID string) (
	[]app.Component, error) {
	g, cmps, err := h.getInvertedDependencyGraphByConfigVersion(ctx, appID, configVersion)

	if err != nil {
		return nil, fmt.Errorf("unable to get graph: %w", err)
	}

	return h.getInvertedDependentComponentsFromComponents(ctx, &g, &cmps, compRootID)
}

// getInvertedDependentComponentsFromComponents from list of components, retrieves the components that depend on the given component root ID
func (h *Helpers) getInvertedDependentComponentsFromComponents(ctx context.Context, g *graph.Graph[string, *app.Component], cmps *[]app.Component, compRootID string) ([]app.Component, error) {
	cmpsById := make(map[string]app.Component)
	for _, c := range *cmps {
		cmpsById[c.ID] = c
	}

	depsCmpIds := make([]string, 0)
	if err := graph.BFS(*g, compRootID, func(compID string) bool {
		if compID == compRootID {
			return false
		}

		if !slices.Contains(depsCmpIds, compID) {
			depsCmpIds = append(depsCmpIds, compID)
		}

		return false
	}); err != nil {
		return nil, fmt.Errorf("unable to build app graph: %w", err)
	}

	depCmps := make([]app.Component, 0)
	for _, id := range depsCmpIds {
		comp, ok := cmpsById[id]
		if !ok {
			return nil, fmt.Errorf("unable to get component")
		}
		depCmps = append(depCmps, comp)
	}

	return depCmps, nil
}
