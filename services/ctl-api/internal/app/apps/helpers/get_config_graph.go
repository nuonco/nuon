package helpers

import (
	"context"

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
	visitedComps := make(map[string]struct{}, 0)

	for _, ccc := range cfg.ComponentConfigConnections {
		ccc.Component.Type = ccc.Type

		visitedComps[ccc.Component.ID] = struct{}{}

		if err := g.AddVertex(&ccc.Component,
			graph.VertexAttribute("name", ccc.Component.Name),
			graph.VertexAttribute("label", ccc.Component.Name),
			graph.VertexAttribute("type", string(ccc.Type)),
			graph.VertexAttribute("color", "blue"),
		); err != nil {
			return nil, err
		}
	}

	allComps, err := h.GetAppComponentsAtConfigVersion(ctx, cfg.AppID, cfg.Version, cfg.ComponentIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app components")
	}
	missingComps := make([]app.ComponentConfigConnection, 0)
	for _, comp := range allComps {
		if _, visited := visitedComps[comp.ID]; visited {
			continue
		}

		if err := g.AddVertex(&comp,
			graph.VertexAttribute("name", comp.Name),
			graph.VertexAttribute("label", comp.Name),
			graph.VertexAttribute("type", string(comp.Type)),
			graph.VertexAttribute("color", "red"),
		); err != nil {
			return nil, err
		}

		if len(comp.ComponentConfigs) < 1 {
			continue
		}

		missingComps = append(missingComps, comp.ComponentConfigs[0])
	}

	// add all dependencies
	allCfgs := append(cfg.ComponentConfigConnections, missingComps...)
	for _, ccc := range allCfgs {
		for _, dep := range ccc.ComponentDependencyIDs {
			if err := g.AddEdge(dep, ccc.ComponentID,
				graph.EdgeWeight(25),
				graph.EdgeAttribute("color", "red"),
			); err != nil {
				return nil, err
			}
		}
	}

	gr, err := graph.TransitiveReduction(g)
	if err != nil {
		return nil, errors.Wrap(err, "unable to reduce graph")
	}
	_ = gr

	return g, nil
}
