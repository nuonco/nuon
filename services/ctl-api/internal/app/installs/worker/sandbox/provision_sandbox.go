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
func (w *Workflows) ProvisionSandbox(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID
	sandboxMode := sreq.SandboxMode

	install, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: installID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	installRun, err := activities.AwaitCreateSandboxRun(ctx, activities.CreateSandboxRunRequest{
		InstallID: installID,
		RunType:   app.SandboxRunTypeProvision,
	})
	if err != nil {
		return fmt.Errorf("unable to create install: %w", err)
	}
	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			w.updateRunStatus(updateCtx, installRun.ID, app.SandboxRunStatusCancelled, "install sandbox run cancelled")
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
			w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "internal error")
			panic(pan)
		}
	}()

	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusProvisioning, "provisioning")

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

	l.Info("executing provision run")
	err = w.executeSandboxRun(ctx, install, installRun, app.RunnerJobOperationTypeCreate, sandboxMode)
	if err != nil {
		return err
	}
	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusActive, "install resources provisioned")

	l.Info("provision was successful")
	return nil
}
