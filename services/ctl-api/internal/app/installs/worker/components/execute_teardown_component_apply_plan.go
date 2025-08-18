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
func (w *Workflows) ExecuteTeardownComponentApplyPlan(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, sreq.ID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	if sreq.ExecuteTeardownComponentSubSignal.ComponentID == "" {
		return fmt.Errorf("component ID is required")
	}

	installDeploy, err := activities.AwaitGetInstallDeployForApplyStep(ctx, activities.GetInstallDeployForApplyStep{
		InstallWorkflowID: sreq.FlowID,
		ComponentID:       sreq.ExecuteTeardownComponentSubSignal.ComponentID,
	})
	if err != nil {
		w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusError, "unable to get install deploy from previous step")
		return errors.Wrap(err, "unable to get install deploy")
	}

	sreq.DeployID = installDeploy.ID
	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			w.updateDeployStatus(updateCtx, installDeploy.ID, app.InstallDeployStatusCancelled, "teardown cancelled")
		}
	}()

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         sreq.WorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: plugins.TableName(w.db, installDeploy),
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	ctx = cctx.SetLogStreamWorkflowContext(ctx, &installDeploy.LogStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	installDeploy, err = activities.AwaitGetDeployByDeployID(ctx, installDeploy.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install deploy")
	}

	l.Info("executing plan")
	if err := w.execApplyPlan(ctx, install, installDeploy, sreq.FlowStepID, sreq.SandboxMode); err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to deploy")
		return errors.Wrap(err, "unable to execute deploy")
	}
	w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusInactive, "successfully torn down")
	w.updateInstallComponentStatus(ctx, installDeploy.InstallComponentID, app.InstallComponentStatusInactive, "successfully torn down")

	_, err = state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       install.ID,
		TriggeredByID:   installDeploy.ID,
		TriggeredByType: plugins.TableName(w.db, installDeploy),
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}

	return err
}
