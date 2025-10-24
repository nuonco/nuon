package components

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ExecuteDeployComponentSyncImage(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, sreq.ID)
	if err != nil {
		w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusError, "unable to get install from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	var installDeploy *app.InstallDeploy
	if sreq.ExecuteDeployComponentSubSignal.DeployID != "" {
		installDeploy, err = activities.AwaitGetDeployByDeployID(ctx, sreq.ExecuteDeployComponentSubSignal.DeployID)
		if err != nil {
			return errors.Wrap(err, "unable to get deploy")
		}

		sreq.DeployID = installDeploy.ID
	} else {
		componentBuild, err := activities.AwaitGetComponentLatestBuildByComponentID(ctx, sreq.ExecuteDeployComponentSubSignal.ComponentID)
		if err != nil {
			return fmt.Errorf("unable to get component build: %w", err)
		}

		typ := app.InstallDeployTypeSync
		installDeploy, err = activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
			InstallID:   install.ID,
			ComponentID: sreq.ExecuteDeployComponentSubSignal.ComponentID,
			BuildID:     componentBuild.ID,
			Type:        typ,
			WorkflowID:  sreq.FlowID,
		})
		if err != nil {
			return fmt.Errorf("unable to create install deploy: %w", err)
		}
		sreq.DeployID = installDeploy.ID
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         sreq.WorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: plugins.TableName(w.db, installDeploy),
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		DeployID: sreq.DeployID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()

	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("syncing oci artifact")
	if err := w.execSync(ctx, install, installDeploy, sreq.SandboxMode); err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to sync")
		return errors.Wrap(err, "unable to execute sync")
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         sreq.WorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: plugins.TableName(w.db, installDeploy),
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusActive, "finished")
	_, err = state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       install.ID,
		TriggeredByID:   installDeploy.ID,
		TriggeredByType: plugins.TableName(w.db, installDeploy),
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}
	return nil
}
