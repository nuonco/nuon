package activities

import (
	"context"
	"fmt"
	"slices"

	"github.com/dominikbraun/graph"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentDependents struct {
	AppID           string `json:"app_id"`
	ComponentRootID string `json:"component_root_id"`
}

// @temporal-gen activity
func (a *Activities) GetComponentDependents(ctx context.Context, req GetComponentDependents) ([]app.Component, error) {
	g, cmps, err := a.appsHelpers.GetDependencyGraph(ctx, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("unable to get graph: %w", err)
	}

	cmpsById := make(map[string]app.Component)
	for _, c := range cmps {
		cmpsById[c.ID] = c
	}

	depsCmpIds := make([]string, 0)
	if err := graph.BFS(g, req.ComponentRootID, func(compID string) bool {
		// warn: g.Vertex(compID) is not always returning the correct component
		if compID == req.ComponentRootID {
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
			return nil, fmt.Errorf("unable to get component: %w", err)
		}
		depCmps = append(depCmps, comp)
	}

	return depCmps, nil
}
