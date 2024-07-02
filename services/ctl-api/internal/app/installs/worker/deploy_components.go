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
	var install app.Install
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		InstallID: installID,
	}, &install); err != nil {
		w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFailed)
		return fmt.Errorf("unable to get install: %w", err)
	}

	if !w.isDeployable(install) {
		// automatically skipping
		w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFailed)
		return nil
	}

	var componentIDs []string
	if err := w.defaultExecGetActivity(ctx, w.acts.GetAppGraph, activities.GetAppGraphRequest{
		AppID: install.AppID,
	}, &componentIDs); err != nil {
		w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFailed)
		return fmt.Errorf("unable to get app graph: %w", err)
	}

	deploys := make([]app.InstallDeploy, 0)
	for _, componentID := range componentIDs {
		var componentBuild app.ComponentBuild
		if err := w.defaultExecGetActivity(ctx, w.acts.GetComponentLatestBuild, activities.GetComponentLatestBuildRequest{
			ComponentID: componentID,
		}, &componentBuild); err != nil {
			w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFailed)
			return fmt.Errorf("unable to get component build: %w", err)
		}

		var installDeploy app.InstallDeploy
		if err := w.defaultExecGetActivity(ctx, w.acts.CreateInstallDeploy, activities.CreateInstallDeployRequest{
			InstallID:   installID,
			ComponentID: componentID,
			BuildID:     componentBuild.ID,
			Signal:      async,
		}, &installDeploy); err != nil {
			w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFailed)
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

	w.writeInstallEvent(ctx, installID, signals.OperationDeployComponents, app.OperationStatusFinished)
	return nil
}
