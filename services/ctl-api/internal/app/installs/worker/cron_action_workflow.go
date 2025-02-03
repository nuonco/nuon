package worker

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

// Run an action workflow config from a cron trigger
type CronActionWorkflowRequest struct {
	InstallID        string             `validate:"required" json:"install_id"`
	ActionWorkflowID string             `validate:"required" json:"action_workflow_id"`
	RunEnvVars       map[string]*string `validate:"required" json:"run_env_vars"`
}

func (w *Workflows) CronActionWorkflow(ctx workflow.Context, req *CronActionWorkflowRequest) error {
	actionWorkflowRun, err := activities.AwaitCreateActionWorkflowRun(ctx, &activities.CreateActionWorkflowRunRequest{
		InstallID:        req.InstallID,
		ActionWorkflowID: req.ActionWorkflowID,
		TriggerType:      app.ActionWorkflowTriggerTypeCron,
		TriggeredByID:    req.ActionWorkflowID,
		TriggeredByType:  "action_workflows",
		RunEnvVars:       req.RunEnvVars,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create action workflow config")
	}

	if err := w.actionWorkflowRun(ctx, req.InstallID, actionWorkflowRun.ID); err != nil {
		return errors.Wrap(err, "unable to create action workflow run")
	}

	return nil
}
