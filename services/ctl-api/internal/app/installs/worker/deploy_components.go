package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) DeployComponents(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID

	w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusStarted)

	l := workflow.GetLogger(ctx)
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	if !w.isDeployable(install) {
		// automatically skipping
		w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFailed)
		return nil
	}

	componentIDs, err := activities.AwaitGetAppInstallGraph(ctx, activities.GetAppInstallGraphRequest{
		AppID:     install.AppID,
		InstallID: install.ID,
	})

	if err != nil {
		w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFailed)
		return fmt.Errorf("unable to get app graph: %w", err)
	}

	deploys := make([]*app.InstallDeploy, 0)
	for _, componentID := range componentIDs {
		componentBuild, err := activities.AwaitGetComponentLatestBuildByComponentID(ctx, componentID)
		if err != nil {
			w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFailed)
			return fmt.Errorf("unable to get component build: %w", err)
		}

		installDeploy, err := activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
			InstallID:   installID,
			ComponentID: componentID,
			BuildID:     componentBuild.ID,
		})

		if err != nil {
			w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFailed)
			return fmt.Errorf("unable to create install deploy: %w", err)
		}

		deploys = append(deploys, installDeploy)
	}

	depDeployErrored := false
	for _, installDeploy := range deploys {
		// NOTE(jm): we make a best effort to deploy all components
		sreq.Type = signals.OperationDeploy
		sreq.DeployID = installDeploy.ID

		if depDeployErrored {
			w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusNoop, "error with depenedent component")
			w.writeDeployEvent(ctx, sreq.DeployID, signals.OperationDeploy, app.OperationStatusNoop)
			continue
		}
		if err := w.AwaitDeploy(ctx, sreq); err != nil {
			l.Error("unable to deploy component", zap.Error(err))
			depDeployErrored = true
		}
	}

	// TODO(sdboyer): is this status unreachable if deployComponents is called with async?
	w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFinished)
	return nil
}
