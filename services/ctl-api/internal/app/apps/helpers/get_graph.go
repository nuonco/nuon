package helpers

import (
	"context"
	"fmt"

	"github.com/dominikbraun/graph"
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
