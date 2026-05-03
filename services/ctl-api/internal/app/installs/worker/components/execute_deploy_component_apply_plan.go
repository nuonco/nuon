package components

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/state/stateregenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

// @temporal-gen-v2 workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ExecuteDeployComponentApplyPlan(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, sreq.ID)
	if err != nil {
		w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusError, "unable to get install from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	installDeploy, err := activities.AwaitGetInstallDeployForApplyStep(ctx, activities.GetInstallDeployForApplyStep{
		InstallWorkflowID: sreq.FlowID,
		ComponentID:       sreq.ID,
	})
	if err != nil {
		w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusError, "unable to get install deploy from previous step")
		return errors.Wrap(err, "unable to get install deploy")
	}

	ctx = cctx.SetLogStreamWorkflowContext(ctx, &installDeploy.LogStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("executing plan")
	if err := w.execApplyPlan(ctx, install, installDeploy, sreq.FlowStepID, sreq.SandboxMode); err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to deploy")
		return errors.Wrap(err, "unable to execute deploy")
	}

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusActive, "finished")
	sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   install.ID,
		OwnerType: "installs",
		QueueName: installshelpers.InstallStateManagerQueueName,
		Signal: &stateregenerate.Signal{
			InstallID:        install.ID,
			Targets:          statemanager.TargetsForHint(statemanager.HintDeployCompleted, sreq.ID),
			ForceAll:         true,
			TriggeredByID:    installDeploy.ID,
			TriggeredByType:  "install_deploys",
			StateGeneratedBy: app.InstallStateGenerateSourceStateManager,
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}
	return nil
}
