package worker

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

// Run an action workflow config from a cron trigger
type CronActionWorkflowRequest struct {
	ID string `validate:"required" json:"id"`
}

func (w *Workflows) CronActionWorkflow(ctx workflow.Context, req *CronActionWorkflowRequest) error {
	installActionWorkflow, err := activities.AwaitGetInstallActionWorkflowByID(ctx, req.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install action workflow")
	}

	runEnvVars := map[string]*string{
		"TRIGGER_TYPE": generics.ToPtr("cron"),
	}

	actionWorkflowRun, err := activities.AwaitCreateActionWorkflowRun(ctx, &activities.CreateActionWorkflowRunRequest{
		InstallActionWorkflowID: installActionWorkflow.ID,
		ActionWorkflowID:        installActionWorkflow.ActionWorkflowID,
		InstallID:               installActionWorkflow.InstallID,
		TriggerType:             app.ActionWorkflowTriggerTypeCron,
		TriggeredByID:           installActionWorkflow.ActionWorkflowID,
		TriggeredByType:         "action_workflows",
		RunEnvVars:              runEnvVars,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create action workflow config")
	}

	if err := w.executeActionWorkflowRun(ctx, installActionWorkflow.InstallID, actionWorkflowRun.ID); err != nil {
		return errors.Wrap(err, "unable to create action workflow run")
	}

	return nil
}
