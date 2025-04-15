package worker

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) TriggerActionWorkflow(ctx workflow.Context, req signals.RequestSignal) error {
	installActionWorkflow, err := activities.AwaitGetInstallActionWorkflowByID(ctx, req.InstallActionWorkflowTrigger.InstallActionWorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get install action workflow")
	}

	actionWorkflowRun, err := activities.AwaitCreateActionWorkflowRun(ctx, &activities.CreateActionWorkflowRunRequest{
		InstallActionWorkflowID: installActionWorkflow.ID,
		ActionWorkflowID:        installActionWorkflow.ActionWorkflowID,
		InstallID:               installActionWorkflow.InstallID,
		TriggerType:             req.InstallActionWorkflowTrigger.TriggerType,
		TriggeredByID:           req.InstallActionWorkflowTrigger.TriggeredByID,
		TriggeredByType:         req.InstallActionWorkflowTrigger.TriggeredByType,
		RunEnvVars:              generics.ToPtrStringMap(req.InstallActionWorkflowTrigger.RunEnvVars),
	})
	if err != nil {
		return errors.Wrap(err, "unable to create action workflow run")
	}

	if err := w.actionWorkflowRun(ctx, installActionWorkflow.InstallID, actionWorkflowRun.ID); err != nil {
		return errors.Wrap(err, "unable to create action workflow run")
	}

	return nil
}
