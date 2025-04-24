package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ExecuteTeardownComponent(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetByInstallID(ctx, sreq.ID)
	if err != nil {
		w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusError, "unable to get install from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	var installDeploy *app.InstallDeploy
	componentBuild, err := activities.AwaitGetComponentLatestBuildByComponentID(ctx, sreq.ExecuteDeployComponentSubSignal.ComponentID)
	if err != nil {
		return fmt.Errorf("unable to get component build: %w", err)
	}

	installDeploy, err = activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
		InstallID:   install.ID,
		ComponentID: sreq.ExecuteTeardownComponentSubSignal.ComponentID,
		BuildID:     componentBuild.ID,
		Teardown:    true,
	})
	if err != nil {
		return fmt.Errorf("unable to create install deploy: %w", err)
	}
	sreq.DeployID = installDeploy.ID

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

	l.Info("performing deploy")
	err = w.doTeardown(ctx, sreq, install)
	if err != nil {
		return errors.Wrap(err, "unable to perform deploy")
	}
	return err
}

func (w *Workflows) doTeardown(ctx workflow.Context, sreq signals.RequestSignal, install *app.Install) error {
	deployID := sreq.DeployID
	sandboxMode := sreq.SandboxMode

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, deployID)
	if err != nil {
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "unable to get install deploy from database")
		return fmt.Errorf("unable to get install deploy: %w", err)
	}

	err = w.pollForDeployableBuild(ctx, deployID, installDeploy.ComponentBuild)
	if err != nil {
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusNoop, "build is not deployable")
		return nil
	}

	if err := w.execSync(ctx, install, installDeploy, sandboxMode); err != nil {
		return w.errorResponse(ctx, sreq, deployID, installDeploy.InstallComponentID, "error syncing", err)
	}

	if err := w.execDeploy(ctx, install, installDeploy, sandboxMode); err != nil {
		return w.errorResponse(ctx, sreq, deployID, installDeploy.InstallComponentID, "error deploying", err)
	}

	w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusNoop, "build is not deployable")

	finalDeployStatus := app.InstallDeployStatusInactive
	finalComponentStatus := app.InstallComponentStatusInactive

	w.updateDeployStatusWithoutStatusSync(ctx, deployID, finalDeployStatus, "successfully torn down")
	w.updateInstallComponentStatus(ctx, installDeploy.InstallComponentID, finalComponentStatus, "successfully torn down")

	return nil
}
