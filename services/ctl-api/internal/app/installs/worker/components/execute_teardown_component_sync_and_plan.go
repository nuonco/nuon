package components

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
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
func (w *Workflows) ExecuteTeardownComponentSyncAndPlan(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, sreq.ID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	if sreq.ExecuteTeardownComponentSubSignal.ComponentID == "" {
		return fmt.Errorf("component ID is required")
	}

	var installDeploy *app.InstallDeploy
	latestDeploy, err := activities.AwaitGetLatestDeploy(ctx, activities.GetLatestDeployRequest{
		ComponentID: sreq.ExecuteTeardownComponentSubSignal.ComponentID,
		InstallID:   install.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// no op, because no deploys exist
			return nil
		}

		return fmt.Errorf("unable to get component build: %w", err)
	}

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, latestDeploy.ComponentBuildID)
	if err != nil {
		return errors.Wrap(err, "unable to get build")
	}

	installDeploy, err = activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
		InstallID:   install.ID,
		ComponentID: sreq.ExecuteTeardownComponentSubSignal.ComponentID,
		BuildID:     build.ID,
		Type:        app.InstallDeployTypeTeardown,
		WorkflowID:  sreq.FlowID,
	})
	if err != nil {
		return fmt.Errorf("unable to create install deploy: %w", err)
	}
	sreq.DeployID = installDeploy.ID
	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			w.updateDeployStatusWithoutStatusSync(updateCtx, installDeploy.ID, app.InstallDeployStatusCancelled, "teardown cancelled")
		}
	}()

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

	installDeploy, err = activities.AwaitGetDeployByDeployID(ctx, installDeploy.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install deploy")
	}

	shouldTeardown := true
	comp, err := activities.AwaitGetComponentByComponentID(ctx, installDeploy.InstallComponent.ComponentID)
	if err != nil {
		return errors.Wrap(err, "unable to get component")
	}
	if generics.SliceContains(comp.Type, []app.ComponentType{}) {
		l.Info("nothing to teardown")
		shouldTeardown = false
	}

	if shouldTeardown {
		l.Info("performing component teardown")
		err = w.doTeardown(ctx, sreq, install)
		if err != nil {
			return errors.Wrap(err, "unable to perform deploy")
		}
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

	err = w.pollForDeployableBuild(ctx, deployID, installDeploy.ComponentBuildID)
	if err != nil {
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusNoop, "build is not deployable")
		return nil
	}

	if err := w.execSync(ctx, install, installDeploy, sandboxMode); err != nil {
		return w.errorResponse(ctx, sreq, deployID, installDeploy.InstallComponentID, "error syncing", err)
	}

	if err := w.execPlan(ctx, install, installDeploy, sreq.FlowStepID, sandboxMode); err != nil {
		return w.errorResponse(ctx, sreq, deployID, installDeploy.InstallComponentID, "error deploying", err)
	}

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPendingApproval, "pending-approval")

	return nil
}
