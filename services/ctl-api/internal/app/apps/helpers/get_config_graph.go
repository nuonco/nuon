package helpers

import (
	"context"
	"slices"

	"github.com/dominikbraun/graph"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// GetConfigGraph builds a directed acyclic graph of component dependencies and returns both the graph
// and a sorted list of components in deployment order (dependencies first)
func (h *Helpers) GetConfigGraph(ctx context.Context, cfg *app.AppConfig) (graph.Graph[string, *app.Component], error) {
	g := graph.New(componentHash,
		graph.Directed(),
		graph.PreventCycles(),
		graph.Rooted(),
		graph.Acyclic())

	// add all components to the config here
	for _, ccc := range cfg.ComponentConfigConnections {
		ccc.Component.Type = ccc.Type

		if err := g.AddVertex(&ccc.Component,
			graph.VertexAttribute("name", ccc.Component.Name),
			graph.VertexAttribute("type", string(ccc.Type)),
		); err != nil {
			return nil, err
		}
	}

	// add all dependencies
	for _, ccc := range cfg.ComponentConfigConnections {
		for _, dep := range ccc.ComponentDependencyIDs {
			if err := g.AddEdge(dep, ccc.ComponentID); err != nil {
				return nil, err
			}
		}
	}

	return g, nil
}

func (h *Helpers) GetConfigDefaultComponentOrder(ctx context.Context, cfg *app.AppConfig) ([]string, error) {
	grph, err := h.GetConfigGraph(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config graph")
	}

	diff := func(a, b string) bool {
		aNode, _ := grph.Vertex(a)
		bNode, _ := grph.Vertex(b)

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

func (h *Helpers) GetConfigReverseDefaultComponentOrder(ctx context.Context, cfg *app.AppConfig) ([]string, error) {
	order, err := h.GetConfigDefaultComponentOrder(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component order")
	}

	slices.Reverse(order)

	return order, nil
}
