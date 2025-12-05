package components

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
func (w *Workflows) ExecuteDeployComponentSyncAndPlan(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, sreq.ID)
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, sreq.DeployID, app.InstallDeployStatusError, "unable to get install from database")
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

		installDeploy, err = activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
			InstallID:   install.ID,
			ComponentID: sreq.ExecuteDeployComponentSubSignal.ComponentID,
			BuildID:     componentBuild.ID,
			Type:        app.InstallDeployTypeApply,
			WorkflowID:  sreq.FlowID,
		})
		if err != nil {
			return fmt.Errorf("unable to create install deploy: %w", err)
		}
		sreq.DeployID = installDeploy.ID
	}

	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			w.updateDeployStatusWithoutStatusSync(updateCtx, installDeploy.ID, app.InstallDeployStatusCancelled, "deploy cancelled")
		}
	}()

	err = w.pollForDeployableBuild(ctx, installDeploy.ID, installDeploy.ComponentBuildID)
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusNoop, "build is not deployable")
		return errors.Wrap(err, "failed to poll for build")
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         sreq.WorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: plugins.TableName(w.db, installDeploy),
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	defer func() {
		if pan := recover(); pan != nil {
			w.updateDeployStatusWithoutStatusSync(ctx, sreq.DeployID, app.InstallDeployStatusError, "internal error")
			panic(pan)
		}
	}()

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		DeployID: sreq.DeployID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	// NOTE: not closed so we can re-use this log stream
	// defer func() {
	// 	activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	// }()

	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("syncing oci artifact")
	if err := w.execSync(ctx, install, installDeploy, sreq.SandboxMode); err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to sync")
		return errors.Wrap(err, "unable to execute sync")
	}

	l.Info("creating plan")
	if err := w.execPlan(ctx, install, installDeploy, sreq.FlowStepID, sreq.SandboxMode); err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to deploy")
		return errors.Wrap(err, "unable to execute deploy")
	}

	w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusPendingApproval, "pending-approval")
	return nil
}
