package activities

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateActionWorkflowRunRequest struct {
	InstallID        string `json:"install_id" validate:"required"`
	ActionWorkflowID string `json:"action_workflow_id" validate:"required"`

	TriggerType     app.ActionWorkflowTriggerType `json:"trigger_type" validate:"required"`
	TriggeredByID   string                        `json:"triggered_by_id"`
	TriggeredByType string                        `json:"triggered_by_type"`

	RunEnvVars map[string]*string `json:"run_env_vars"`
}

// @temporal-gen activity
func (a *Activities) CreateActionWorkflowRun(ctx context.Context, req *CreateActionWorkflowRunRequest) (*app.InstallActionWorkflowRun, error) {
	return a.createActionWorkflowRun(ctx,
		req.InstallID,
		req.ActionWorkflowID,
		req.TriggerType,
		req.RunEnvVars,
		req.TriggeredByID,
		req.TriggeredByType,
	)
}

func (a *Activities) createActionWorkflowRun(ctx context.Context,
	installID,
	actionWorkflowID string,
	triggerType app.ActionWorkflowTriggerType,
	runEnvVars map[string]*string,
	triggeredByID string,
	triggeredByType string,
) (*app.InstallActionWorkflowRun, error) {
	cfg, err := a.getActionWorkflowLatestConfig(ctx, actionWorkflowID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get latest action workflow config")
	}

	steps := make([]app.InstallActionWorkflowRunStep, 0)
	for _, step := range cfg.Steps {
		steps = append(steps, app.InstallActionWorkflowRunStep{
			Status: app.InstallActionWorkflowRunStepStatusPending,
			StepID: step.ID,
		})
	}

	newRun := app.InstallActionWorkflowRun{
		InstallID:              installID,
		ActionWorkflowConfigID: cfg.ID,
		TriggerType:            triggerType,
		Status:                 app.InstallActionRunStatusQueued,
		StatusDescription:      "Queued",
		Steps:                  steps,
		RunEnvVars:             pgtype.Hstore(runEnvVars),
		TriggeredByID:          triggeredByID,
		TriggeredByType:        triggeredByType,
	}

	res := a.db.WithContext(ctx).
		Create(&newRun)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create action workflow")
	}

	return &newRun, nil
}
