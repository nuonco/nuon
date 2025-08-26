package sandbox

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
// @execution-timeout 30m
func (w *Workflows) DeprovisionSandboxApplyPlan(ctx workflow.Context, sreq signals.RequestSignal) error {
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

	ctx = cctx.SetLogStreamWorkflowContext(ctx, &installRun.LogStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, installRun.LogStream.ID)
	}()

	l.Info("executing plan")
	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusDeprovisioned, "successfully deprovisioned")
	if err := w.executeApplyPlan(ctx, install, installRun, sreq.FlowStepID, sreq.SandboxMode); err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "job did not succeed")
		return errors.Wrap(err, "unable to execute deploy")
	}

	_, err = state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       install.ID,
		TriggeredByID:   installRun.ID,
		TriggeredByType: plugins.TableName(w.db, installRun),
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}

	return nil
}
