package worker

import (
	"fmt"
	"slices"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) shouldTeardownInstallComponent(ctx workflow.Context, installID, compID string) (bool, error) {
	installComponent, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
		InstallID:   installID,
		ComponentID: compID,
	})
	if err != nil {
		return false, fmt.Errorf("unable to get install component: %w", err)
	}

	if len(installComponent.InstallDeploys) < 1 {
		return false, nil
	}

	//if installComponent.InstallDeploys[0].Status != string(StatusActive) {
	//return false, nil
	//}

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

func (w *Workflows) teardownComponents(ctx workflow.Context, installID string, sandboxMode, async bool) error {
	l := workflow.GetLogger(ctx)
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	// reasons we should not try to teardown components
	if !w.shouldTeardownComponents(install) {
		return nil
	}

	componentIDs, err := activities.AwaitGetAppGraphByAppID(ctx, install.AppID)
	if err != nil {
		return fmt.Errorf("unable to get app graph: %w", err)
	}

	// NOTE(jm): it would probably be better, long term to have a proper way of inverting the graph and walking it
	// in reverse, but for now, this is the only place we need to do so, so it is just localized here.
	slices.Reverse(componentIDs)

	deploys := make([]*app.InstallDeploy, 0)
	for _, compID := range componentIDs {
		shouldTeardown, err := w.shouldTeardownInstallComponent(ctx, installID, compID)
		if err != nil {
			return fmt.Errorf("unable to verify if component should be torn down: %w", err)
		}

		if !shouldTeardown {
			continue
		}

		componentBuild, err := activities.AwaitGetComponentLatestBuildByComponentID(ctx, compID)
		if err != nil {
			return fmt.Errorf("unable to get latest component build: %w", err)
		}

		installDeploy, err := activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
			InstallID: installID, ComponentID: compID,
			BuildID:  componentBuild.ID,
			Teardown: true,
			Signal:   async,
		})
		if err != nil {
			return fmt.Errorf("unable to create install deploy: %w", err)
		}

		deploys = append(deploys, installDeploy)
	}

	if async {
		return nil
	}
	for _, installDeploy := range deploys {
		// NOTE(jm): we make a best effort to teardown all components
		if err := w.deploy(ctx, installID, installDeploy.ID, sandboxMode); err != nil {
			l.Error("unable to teardown component", zap.Error(err))
		}
	}

	return nil
}
