package worker

import (
	"errors"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) shouldTeardownInstallComponent(ctx workflow.Context, installID, compID string) (bool, error) {
	installComponent, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
		InstallID:   installID,
		ComponentID: compID,
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("unable to get install component: %w", err)
	}

	if installComponent == nil {
		return false, nil
	}

	if len(installComponent.InstallDeploys) < 1 {
		return false, nil
	}

	lastInstallDeploy := installComponent.InstallDeploys[0]
	if lastInstallDeploy.Status == app.InstallDeployStatusInactive {
		return false, nil
	}

	return true, nil
}

func (w *Workflows) shouldTeardownComponents(install *app.Install) bool {
	if len(install.InstallSandboxRuns) < 1 {
		return false
	}

	lastRun := install.InstallSandboxRuns[0]
	if (lastRun.RunType == app.SandboxRunTypeProvision ||
		lastRun.RunType == app.SandboxRunTypeReprovision) &&
		lastRun.Status == app.SandboxRunStatusActive {
		return true
	}

	return false
}
