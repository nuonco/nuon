package sandbox

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
func (w *Workflows) ProvisionSandboxPlan(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetInstallForSandboxBySandboxID(ctx, sreq.ID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}
	sandboxMode := sreq.SandboxMode

	installRun, err := activities.AwaitCreateSandboxRun(ctx, activities.CreateSandboxRunRequest{
		InstallID:  install.ID,
		RunType:    app.SandboxRunTypeProvision,
		WorkflowID: sreq.FlowID,
	})
	if err != nil {
		return fmt.Errorf("unable to create install: %w", err)
	}
	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			w.updateRunStatusWithoutStatusSync(updateCtx, installRun.ID, app.SandboxRunStatusCancelled, "install sandbox run cancelled")
		}
	}()

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         sreq.WorkflowStepID,
		StepTargetID:   installRun.ID,
		StepTargetType: plugins.TableName(w.db, installRun),
	}); err != nil {
		return errors.Wrap(err, "unable to update install action workflow")
	}

	defer func() {
		if pan := recover(); pan != nil {
			w.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "internal error")
			panic(pan)
		}
	}()

	w.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatus(app.InstallDeployStatusPlanning), "planning")

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		SandboxRunID: installRun.ID,
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

	l.Info("executing provision plan")
	err = w.executeSandboxPlan(ctx, install, installRun, sreq.FlowStepID, sandboxMode)
	if err != nil {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
		return err
	}

	w.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunPendingApproval, "auto skipped")
	l.Info("provision plan was successful")
	return nil
}
