package sandbox

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @execution-timeout 30m
func (w *Workflows) ReprovisionSandboxApplyPlan(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetInstallForSandboxBySandboxID(ctx, sreq.ID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	installRun, err := activities.AwaitGetInstallSandboxRunForApplyStep(ctx, activities.GetInstallSandboxRunForApplyStep{
		InstallWorkflowID: sreq.FlowID,
		InstallID:         install.ID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get install deploy")
	}

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		DeployID: sreq.DeployID,
	})
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "internal error")
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

	l.Info("executing plan")
	if err := w.executeApplyPlan(ctx, install, installRun, sreq.FlowStepID, sreq.SandboxMode); err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "job did not succeed")
		return errors.Wrap(err, "unable to execute deploy")
	}

	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusActive, "successfully reprovisioned")
	return nil
}
