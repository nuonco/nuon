package worker

import (
	"fmt"
	"sort"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) teardownComponents(ctx workflow.Context, installID string, sandboxMode, async bool) error {
	l := workflow.GetLogger(ctx)
	var install app.Install
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		InstallID: installID,
	}, &install); err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	var componentIDs []string
	if err := w.defaultExecGetActivity(ctx, w.acts.GetAppGraph, activities.GetAppGraphRequest{
		AppID: install.AppID,
	}, &componentIDs); err != nil {
		return fmt.Errorf("unable to get app graph: %w", err)
	}

	// NOTE(jm): it would probably be better, long term to have a proper way of inverting the graph and walking it
	// in reverse, but for now, this is the only place we need to do so, so it is just localized here.
	sort.Sort(sort.Reverse(sort.StringSlice(componentIDs)))

	deploys := make([]app.InstallDeploy, 0)
	for _, compID := range componentIDs {

		var componentBuild app.ComponentBuild
		if err := w.defaultExecGetActivity(ctx, w.acts.GetComponentLatestBuild, activities.GetComponentLatestBuildRequest{
			ComponentID: compID,
		}, &componentBuild); err != nil {
			return fmt.Errorf("unable to create component build: %w", err)
		}

		var installDeploy app.InstallDeploy
		if err := w.defaultExecGetActivity(ctx, w.acts.CreateInstallDeploy, activities.CreateInstallDeployRequest{
			InstallID:   installID,
			ComponentID: compID,
			BuildID:     componentBuild.ID,
			Teardown:    true,
			Signal:      async,
		}, &installDeploy); err != nil {
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
