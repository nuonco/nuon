package helpers

import (
	"context"
	"slices"

	"github.com/dominikbraun/graph"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) getDeployOrderFromGraph(ctx context.Context, grph graph.Graph[string, *app.Component]) ([]string, error) {
	diff := func(a, b string) bool {
		aNode, _ := grph.Vertex(a)
		bNode, _ := grph.Vertex(b)

		if len(aNode.Dependencies) != len(bNode.Dependencies) {
			return len(aNode.Dependencies) < len(bNode.Dependencies)
		}

		typeOrder := map[app.ComponentType]int{
			"external_image":   0,
			"docker_build":     1,
			"terraform_module": 2,
			"helm_chart":       3,
			"job":              4,
		}
		aType := typeOrder[aNode.Type]
		bType := typeOrder[bNode.Type]

		if aType == bType {
			return aNode.Name < bNode.Name
		}

		return aType < bType
	}

	// Perform topological sort
	order, err := graph.StableTopologicalSort(grph, diff)
	if err != nil {
		return nil, errors.Wrap(err, "unable to perform topological sort")
	}

	return order, nil
}

func (h *Helpers) GetConfigDefaultComponentOrder(ctx context.Context, cfg *app.AppConfig) ([]string, error) {
	grph, err := h.GetConfigGraph(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config graph")
	}

	return h.getDeployOrderFromGraph(ctx, grph)
}

func (h *Helpers) GetConfigReverseDefaultComponentOrder(ctx context.Context, cfg *app.AppConfig) ([]string, error) {
	grph, err := h.GetConfigGraph(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config graph")
	}

	comps, err := h.getDeployOrderFromGraph(ctx, grph)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get deploy order from graph")
	}

	slices.Reverse(comps)

	return comps, nil
}

func (h *Helpers) GetConfigComponentDeployOrder(ctx context.Context, cfg *app.AppConfig, compID string) ([]string, error) {
	grph, err := h.GetConfigGraph(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config graph")
	}

	visitedCmps := make([]string, 0)
	err = graph.BFSWithDepth(grph, compID, func(id string, depth int) bool {
		visitedCmps = append(visitedCmps, id)
		return false
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get subgraph")
	}

	sortedComps, err := h.getDeployOrderFromGraph(ctx, grph)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get deploy order from graph")
	}

	sortedVisitedCmps := make([]string, 0, len(visitedCmps))
	for _, id := range sortedComps {
		for _, visitedID := range visitedCmps {
			if id == visitedID {
				sortedVisitedCmps = append(sortedVisitedCmps, id)
				break
			}
		}
	}

	return sortedVisitedCmps, nil
}

func (h *Helpers) GetReverseConfigComponentDeployOrder(ctx context.Context, cfg *app.AppConfig, compID string) ([]string, error) {
	comps, err := h.GetConfigComponentDeployOrder(ctx, cfg, compID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component deploy order")
	}

	slices.Reverse(comps)
	return comps, nil
}
