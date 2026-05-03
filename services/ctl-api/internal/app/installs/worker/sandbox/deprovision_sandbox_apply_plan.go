package sandbox

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	workerstate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// @temporal-gen-v2 workflow
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
	if err := w.executeApplyPlan(ctx, install, installRun, sreq.FlowStepID, sreq.SandboxMode); err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "job did not succeed")
		return errors.Wrap(err, "unable to execute deploy")
	}
	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusDeprovisioned, "successfully deprovisioned")

	workerstate.AwaitHintStateManager(ctx, &workerstate.HintStateManagerRequest{
		InstallID:       install.ID,
		HintType:        statemanager.HintSandboxDeprovisioned,
		TriggeredByID:   installRun.ID,
		TriggeredByType: app.InstallStateGenerateSourceStateManager,
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}

	return nil
}
