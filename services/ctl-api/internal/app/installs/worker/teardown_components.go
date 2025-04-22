package worker

import (
	"errors"
	"fmt"
	"slices"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
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

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) TeardownComponents(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID

	l := workflow.GetLogger(ctx)
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	// fail all queued deploys
	if err := activities.AwaitFailQueuedDeploysByInstallID(ctx, installID); err != nil {
		return fmt.Errorf("unable to fail queued install: %w", err)
	}

	// reasons we should not try to teardown components
	if !w.shouldTeardownComponents(install) {
		return nil
	}

	componentIDs, err := activities.AwaitGetAppInstallGraph(ctx, activities.GetAppInstallGraphRequest{
		AppID:     install.AppID,
		InstallID: install.ID,
	})
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
		})
		if err != nil {
			return fmt.Errorf("unable to create install deploy: %w", err)
		}

		deploys = append(deploys, installDeploy)
	}

	for _, installDeploy := range deploys {
		sreq.Type = signals.OperationDeploy
		sreq.DeployID = installDeploy.ID

		if err := w.Deploy(ctx, sreq); err != nil {
			l.Error("unable to teardown component, continuing to the next deploy", zap.Error(err))
		}
	}

	w.evClient.Send(ctx, install.ID, &signals.Signal{
		Type:        signals.OperationDeprovision,
		ForceDelete: sreq.ForceDelete,
	})

	return nil
}
