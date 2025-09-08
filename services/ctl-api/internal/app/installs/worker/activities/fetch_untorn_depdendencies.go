package activities

import (
	"context"
	"fmt"
	"go.temporal.io/sdk/temporal"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type FetchUntornDependenciesRequest struct {
	ComponentRootID string `json:"component_root_id" validate:"required"`
	InstallID       string `json:"install_id" validate:"required"`
}

// @temporal-gen activity
func (a *Activities) FetchUntornDependencies(ctx context.Context, req FetchUntornDependenciesRequest) ([]string, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(
			"unable to get install",
			"unable to get install",
			err)
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
			return depIds, fmt.Errorf("error getting latest deploy: %w", err)
		}

		if installCmpDeploy == nil {
			continue
		}

		if app.InstallDeployStatus(app.InstallDeployStatus(*install.ComponentStatuses[dep.ID])) != app.InstallDeployStatusActive && installCmpDeploy.Type != app.InstallDeployTypeTeardown {
			depIds = append(depIds, dep.ID)
		}
	}

	return depIds, nil
}
