package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type FetchUntornDependenciesRequest struct {
	ComponentRootID string `json:"component_root_id"`
	InstallID       string `json:"install_id"`
}

// @await-gen
func (a *Activities) FetchUntornDependencies(ctx context.Context, req FetchUntornDependenciesRequest) ([]string, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	depIds := make([]string, 0)
	depComponents, err := a.appsHelpers.GetInvertedDependentComponents(ctx, install.App.ID, req.ComponentRootID)
	if err != nil {
		return depIds, fmt.Errorf("unable to GetInvertedDependentComponents: %w", err)
	}

	for _, dep := range depComponents {
		// skip if dependency has no install component
		if _, ok := install.ComponentStatuses[dep.ID]; !ok {
			continue
		}

		installCmpDeploy, err := a.getLatestDeploy(ctx, req.InstallID, dep.ID)
		if err != nil {
			return depIds, fmt.Errorf("er")
		}

		if installCmpDeploy == nil {
			continue
		}

		if app.InstallDeployStatus(app.InstallDeployStatus(*install.ComponentStatuses[dep.ID])) != app.InstallDeployStatusOK && installCmpDeploy.Type != app.InstallDeployTypeTeardown {
			depIds = append(depIds, dep.ID)
		}
	}

	return depIds, nil
}
