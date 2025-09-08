package helpers

import (
	"context"
	"fmt"

	"github.com/dominikbraun/graph"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

func componentHash(a *app.Component) string {
	return a.ID
}

type errComponentVertex struct {
	err    error
	compID string
}

func (e *errComponentVertex) Error() string {
	return fmt.Errorf("unable to add component to graph: %w", e.err).Error()
}

func (e *errComponentVertex) Unwrap() error {
	return e.err
}

type errComponentEdge struct {
	err          error
	compID       string
	dependencyID string
}

func (e *errComponentEdge) Error() string {
	return fmt.Errorf("unable to add component edge to graph: %w", e.err).Error()
}

func (e *errComponentEdge) Unwrap() error {
	return e.err
}

func (h *Helpers) OrderComponentsByDep(ctx context.Context, components []app.Component) ([]app.Component, error) {
	if len(components) <= 1 {
		return components, nil
	}

	g := graph.New(componentHash,
		graph.Directed(),
		graph.PreventCycles(),
		graph.Rooted(),
		graph.Acyclic())

	for _, comp := range components {
		c := &comp
		if err := g.AddVertex(c); err != nil {
			return nil, err
		}
	}

	rootIDs := make([]string, 0)
	for _, comp := range components {
		if len(comp.Dependencies) < 1 {
			rootIDs = append(rootIDs, comp.ID)
			continue
		}

		for _, dep := range comp.Dependencies {
			if err := g.AddEdge(dep.ID, comp.ID); err != nil {
				return nil, err
			}
		}
	}

	if len(rootIDs) < 1 {
		// NOTE(jm): this should never happen, as that would require a cycled graph, otherwise.
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("at least one component should have 0 dependencies"),
			Description: "at least one component should have 0 dependencies",
		}
	}

	cmpsById := make(map[string]app.Component)
	for _, c := range components {
		cmpsById[c.ID] = c
	}

	orderedComponents := make([]app.Component, 0)
	for _, rootID := range rootIDs {
		if err := graph.BFS(g, rootID, func(compID string) bool {
			orderedComponents = append(orderedComponents, cmpsById[compID])
			return false
		}); err != nil {
			return nil, fmt.Errorf("unable to build app graph: %w", err)
		}
	}

	return orderedComponents, nil
}

func (h *Helpers) GetDependencyGraph(ctx context.Context, appID string) (graph.Graph[string, *app.Component], []app.Component, error) {
	a := app.App{}
	res := h.db.WithContext(ctx).
		Preload("Org").
		Preload("Components").
		Preload("Components.Dependencies").
		Preload("Components.ComponentConfigs").
		Preload("Components.ComponentConfigs.ComponentBuilds").
		First(&a, "id = ?", appID)
	if res.Error != nil {
		return nil, nil, fmt.Errorf("unable to get app with id %s: %w", appID, res.Error)
	}

	g := graph.New(componentHash,
		graph.Directed(),
		graph.PreventCycles(),
		graph.Rooted(),
		graph.Acyclic())

	for _, comp := range a.Components {
		c := &comp
		if err := g.AddVertex(c); err != nil {
			return nil, nil, err
		}
	}

	for _, comp := range a.Components {
		for _, dep := range comp.Dependencies {
			// edge assignment should be comp.ID -> dep.ID in order to BFS search all dependencies
			if err := g.AddEdge(comp.ID, dep.ID); err != nil {
				return nil, nil, err
			}
		}
	}

	// TODO: investigate why g.Vertex(comp.ID) is not returning the correct component
	// for now returning all components for a lookup
	return g, a.Components, nil
}

func (h *Helpers) GetInvertedDependencyGraph(ctx context.Context, appID string) (graph.Graph[string, *app.Component], []app.Component, error) {
	a := app.App{}
	res := h.db.WithContext(ctx).
		Preload("Org").
		Preload("Components").
		Preload("Components.Dependencies").
		Preload("Components.ComponentConfigs").
		Preload("Components.ComponentConfigs.ComponentBuilds").
		First(&a, "id = ?", appID)
	if res.Error != nil {
		return nil, nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return h.getInvertedDependencyGraphFromComponents(ctx, &a.Components)
}

// getInvertedDependencyGraphByConfigVersion builds a graph from components at a specific config version.
func (h *Helpers) getInvertedDependencyGraphByConfigVersion(ctx context.Context, appID string, cfg *app.AppConfig, configVersion int) (graph.Graph[string, *app.Component], []app.Component, error) {
	comps, err := h.GetAppComponentsAtConfigVersion(ctx, appID, configVersion, cfg.ComponentIDs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get app components at config version")
	}

	return h.getInvertedDependencyGraphFromComponents(ctx, &comps)
}

// getInvertedDependencyuuGraphFromComponents builds a inverted dependency graph from input list of components.
func (h *Helpers) getInvertedDependencyGraphFromComponents(ctx context.Context, comps *[]app.Component) (graph.Graph[string, *app.Component], []app.Component, error) {
	g := graph.New(componentHash,
		graph.Directed(),
		graph.PreventCycles(),
		graph.Rooted(),
		graph.Acyclic())

	for _, comp := range *comps {
		c := &comp
		if err := g.AddVertex(c); err != nil {
			return nil, nil, err
		}
	}

	for _, comp := range *comps {
		for _, dep := range comp.Dependencies {
			// edge assignment should be dep.ID -> comp.ID in order to BFS search and fetch all dependents of this component
			if err := g.AddEdge(dep.ID, comp.ID); err != nil {
				return nil, nil, err
			}
		}
	}

	return g, *comps, nil
}

func (h *Helpers) GetGraph(ctx context.Context, appID string) (graph.Graph[string, *app.Component], []string, error) {
	a := app.App{}
	res := h.db.WithContext(ctx).
		Preload("Org").
		Preload("Components").
		Preload("Components.Dependencies").
		Preload("Components.ComponentConfigs").
		Preload("Components.ComponentConfigs.ComponentBuilds").
		First(&a, "id = ?", appID)
	if res.Error != nil {
		return nil, nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	g := graph.New(componentHash,
		graph.Directed(),
		graph.PreventCycles(),
		graph.Rooted(),
		graph.Acyclic())

	for _, comp := range a.Components {
		c := &comp
		if err := g.AddVertex(c); err != nil {
			return nil, nil, err
		}
	}

	rootIDs := make([]string, 0)
	for _, comp := range a.Components {
		if len(comp.Dependencies) < 1 {
			rootIDs = append(rootIDs, comp.ID)
			continue
		}

		for _, dep := range comp.Dependencies {
			if err := g.AddEdge(dep.ID, comp.ID); err != nil {
				return nil, nil, err
			}
		}
	}

	if len(rootIDs) < 1 {
		// NOTE(jm): this should never happen, as that would require a cycled graph, otherwise.
		return nil, nil, stderr.ErrUser{
			Err:         fmt.Errorf("at least one component should have 0 dependencies"),
			Description: "at least one component should have 0 dependencies",
		}
	}

	return g, rootIDs, nil
}
