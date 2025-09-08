package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type FetchInactiveDependenciesRequest struct {
	ComponentRootID string `json:"component_root_id" validate:"required"`
	InstallID       string `json:"install_id" validate:"required"`
}

// @temporal-gen activity
func (a *Activities) FetchInactiveDependencies(ctx context.Context, req FetchInactiveDependenciesRequest) ([]string, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	inactiveDepIds := make([]string, 0)
	depComponents, err := a.appsHelpers.GetDependentComponents(ctx, install.App.ID, req.ComponentRootID)
	if err != nil {
		return inactiveDepIds, fmt.Errorf("unable to GetDependentComponents: %w", err)
	}

	for _, dep := range depComponents {
		if _, ok := install.ComponentStatuses[dep.ID]; !ok {
			continue
		}

		if app.InstallDeployStatus(app.InstallDeployStatus(*install.ComponentStatuses[dep.ID])) != app.InstallDeployStatusActive {
			inactiveDepIds = append(inactiveDepIds, dep.ID)
		}
	}

	return inactiveDepIds, nil
}
