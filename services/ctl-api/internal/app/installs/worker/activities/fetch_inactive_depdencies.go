package activities

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/dominikbraun/graph"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type FetchInactiveDependenciesRequest struct {
	ComponentRootID string `json:"component_root_id"`
	InstallID       string `json:"install_id"`
}

// @await-gen
func (a *Activities) FetchInactiveDependencies(ctx context.Context, req FetchInactiveDependenciesRequest) ([]string, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	inactiveDepIds := make([]string, 0)
	depComponents, err := a.getComponentDependents(ctx, install.App.ID, req.ComponentRootID)
	if err != nil {
		return inactiveDepIds, fmt.Errorf("unable to getComponentDependees: %w", err)
	}

	for _, dep := range depComponents {
		// possible for a stale app component to not be part of the install ignore
		if _, ok := install.ComponentStatuses[dep.ID]; !ok {
			fmt.Println("dep not in install.ComponentStatuses continue")
			continue
		}

		if app.InstallDeployStatus(app.InstallDeployStatus(*install.ComponentStatuses[dep.ID])) != app.InstallDeployStatusOK {
			inactiveDepIds = append(inactiveDepIds, dep.ID)
		}
	}

	if len(inactiveDepIds) > 0 {
		return inactiveDepIds, fmt.Errorf("dependent install components: [%s], inactive", strings.Join(inactiveDepIds, ", "))
	}

	return inactiveDepIds, nil
}

func (a *Activities) getComponentDependents(ctx context.Context, appID string, compRootID string) ([]app.Component, error) {
	g, cmps, err := a.appsHelpers.GetDependencyGraph(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get graph: %w", err)
	}

	cmpsById := make(map[string]app.Component)
	for _, c := range cmps {
		cmpsById[c.ID] = c
	}

	depsCmpIds := make([]string, 0)
	if err := graph.BFS(g, compRootID, func(compID string) bool {
		// warn: g.Vertex(compID) is not always returning the correct component
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
			return nil, fmt.Errorf("unable to get component: %w", err)
		}
		depCmps = append(depCmps, comp)
	}

	return depCmps, nil
}
