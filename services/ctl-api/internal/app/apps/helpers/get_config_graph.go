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

	// get all components and their latest config, and then ensure any components which are not connected to this
	// config use the latest component config to attach them to the graph
	allComps, err := h.GetAppComponentsAndLatestConfigConnection(ctx, cfg.AppID)
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

func (h *Helpers) GetDeployOrderFromGraph(ctx context.Context, grph graph.Graph[string, *app.Component]) ([]string, error) {
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

	return h.GetDeployOrderFromGraph(ctx, grph)
}

func (h *Helpers) GetConfigReverseDefaultComponentOrder(ctx context.Context, cfg *app.AppConfig) ([]string, error) {
	grph, err := h.GetConfigGraph(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config graph")
	}

	comps, err := h.GetDeployOrderFromGraph(ctx, grph)
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

	comps := make([]string, 0)
	err = graph.BFSWithDepth(grph, compID, func(id string, depth int) bool {
		comps = append(comps, id)
		return false
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get subgraph")
	}

	return comps, nil
}

func (h *Helpers) GetReverseConfigComponentDeployOrder(ctx context.Context, cfg *app.AppConfig, compID string) ([]string, error) {
	comps, err := h.GetConfigComponentDeployOrder(ctx, cfg, compID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component deploy order")
	}

	slices.Reverse(comps)
	return comps, nil
}
