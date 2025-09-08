package helpers

import (
	"context"
	"fmt"
	"slices"

	"github.com/dominikbraun/graph"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) GetComponentDependents(ctx context.Context, appCfgID string, compID string) ([]string, error) {
	appCfg, err := h.GetFullAppConfig(ctx, appCfgID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	graph, err := h.GetConfigGraph(ctx, appCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config graph")
	}

	return h.getDependentComponentsByGraph(compID, graph)
}

func (h *Helpers) getDependentComponentsByGraph(compRootID string, g graph.Graph[string, *app.Component]) ([]string, error) {
	depsCmpIDs := make([]string, 0)
	if err := graph.BFS(g, compRootID, func(compID string) bool {
		if compID == compRootID {
			return false
		}

		if !slices.Contains(depsCmpIDs, compID) {
			depsCmpIDs = append(depsCmpIDs, compID)
		}

		return false
	}); err != nil {
		return nil, fmt.Errorf("unable to build app graph: %w", err)
	}

	slices.Reverse(depsCmpIDs)
	return depsCmpIDs, nil
}
