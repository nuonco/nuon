package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) deployComponents(ctx workflow.Context, installID string, sandboxMode, async bool) error {
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

	componentIDs, err := activities.AwaitGetAppGraphByAppID(ctx, install.AppID)
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
			Signal:      async,
		})
		if err != nil {
			w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFailed)
			return fmt.Errorf("unable to create install deploy: %w", err)
		}

		deploys = append(deploys, installDeploy)
	}

	if async {
		return nil
	}

	for _, installDeploy := range deploys {
		// NOTE(jm): we make a best effort to deploy all components
		if err := w.deploy(ctx, installID, installDeploy.ID, sandboxMode); err != nil {
			l.Error("unable to deploy component", zap.Error(err))

			// (rb) stop iterating after first error
			break

		}
	}

	w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFinished)
	return nil
}
