package activities

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateActionWorkflowRunRequest struct {
	ActionWorkflowID        string `json:"action_workflow_id" validate:"required"`
	InstallActionWorkflowID string `json:"install_action_workflow_id" validate:"required"`
	InstallID               string `json:"install_id" validate:"required"`
	InstallWorkflowID       string `json:"install_workflow_id,omitempty"`

	TriggerType     app.ActionWorkflowTriggerType `json:"trigger_type" validate:"required"`
	TriggeredByID   string                        `json:"triggered_by_id"`
	TriggeredByType string                        `json:"triggered_by_type"`

	RunEnvVars map[string]*string `json:"run_env_vars"`
}

// @temporal-gen activity
func (a *Activities) CreateActionWorkflowRun(ctx context.Context, req *CreateActionWorkflowRunRequest) (*app.InstallActionWorkflowRun, error) {
	return a.createActionWorkflowRun(ctx,
		req.ActionWorkflowID,
		req.InstallActionWorkflowID,
		req.InstallID,
		req.InstallWorkflowID,
		req.TriggerType,
		req.TriggeredByID,
		req.TriggeredByType,
		req.RunEnvVars,
	)
}

func (a *Activities) createActionWorkflowRun(ctx context.Context,
	actionWorkflowID string,
	installActionWorkflowID string,
	installID string,
	installWorkflowID string,
	triggerType app.ActionWorkflowTriggerType,
	triggeredByID string,
	triggeredByType string,
	runEnvVars map[string]*string,
) (*app.InstallActionWorkflowRun, error) {
	install, err := a.getInstall(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	cfg, err := a.actionHelpers.GetActionWorkflowConfig(ctx, actionWorkflowID, install.AppConfigID)
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
		InstallActionWorkflowID: installActionWorkflowID,
		InstallID:               installID,
		ActionWorkflowConfigID:  cfg.ID,
		TriggerType:             triggerType,
		Status:                  app.InstallActionRunStatusQueued,
		StatusDescription:       "Queued",
		Steps:                   steps,
		RunEnvVars:              pgtype.Hstore(runEnvVars),
		TriggeredByID:           triggeredByID,
		TriggeredByType:         triggeredByType,
	}

	if installWorkflowID != "" {
		newRun.InstallWorkflowID = generics.ToPtr(installWorkflowID)
	}

	res := a.db.WithContext(ctx).
		Create(&newRun)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create action workflow")
	}

	return &newRun, nil
}
