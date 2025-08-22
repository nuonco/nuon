package sandbox

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

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
func (w *Workflows) ReprovisionSandboxApplyPlan(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetInstallForSandboxBySandboxID(ctx, sreq.ID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	sandboxRun, err := activities.AwaitGetInstallSandboxRunForApplyStep(ctx, activities.GetInstallSandboxRunForApplyStep{
		InstallWorkflowID: sreq.FlowID,
		InstallID:         install.ID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get install deploy")
	}

	ctx = cctx.SetLogStreamWorkflowContext(ctx, &sandboxRun.LogStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, sandboxRun.LogStream.ID)
	}()

	l.Info("executing sandbox apply plan", zap.String("install_run.id", sandboxRun.ID))

	w.updateRunStatus(ctx, sandboxRun.ID, app.SandboxRunStatusReprovisioning, "reprovisioning sandbox")

	err = w.executeApplyPlan(ctx, install, sandboxRun, sreq.FlowStepID, sreq.SandboxMode)
	if err != nil {
		l.Error("error executing sandbox apply plan", zap.String("install_run.id", sandboxRun.ID), zap.Error(err))
		w.updateRunStatus(ctx, sandboxRun.ID, app.SandboxRunStatusError, "job did not succeed")
		return errors.Wrap(err, "unable to execute deploy")
	}
	l.Debug("finished executing sandbox apply plan", zap.String("install_run.id", sandboxRun.ID))

	l.Info("updating install sandbox run status", zap.String("install_run.id", sandboxRun.ID))

	w.updateRunStatus(ctx, sandboxRun.ID, app.SandboxRunStatusActive, "successfully reprovisioned")
	_, err = state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       install.ID,
		TriggeredByID:   sreq.InstallWorkflowID,
		TriggeredByType: plugins.TableName(w.db, sandboxRun),
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}
	return nil
}
