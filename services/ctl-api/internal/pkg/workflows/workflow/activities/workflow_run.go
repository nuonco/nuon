package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateWorkflowRunRequest struct {
	WorkflowID    string              `json:"workflow_id" validate:"required"`
	Type          app.WorkflowRunType `json:"type" validate:"required"`
	TriggerStepID string              `json:"trigger_step_id,omitempty"`
	StartFromIdx  int                 `json:"start_from_idx"`
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowCreateWorkflowRun(ctx context.Context, req CreateWorkflowRunRequest) (*app.WorkflowRun, error) {
	run := &app.WorkflowRun{
		WorkflowID:    req.WorkflowID,
		Type:          req.Type,
		TriggerStepID: req.TriggerStepID,
		StartFromIdx:  req.StartFromIdx,
		Status: app.CompositeStatus{
			Status: app.StatusInProgress,
		},
	}

	if res := a.db.WithContext(ctx).Create(run); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create workflow run")
	}

	return run, nil
}

type UpdateWorkflowRunStatusRequest struct {
	RunID  string              `json:"run_id" validate:"required"`
	Status app.CompositeStatus `json:"status" validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowUpdateWorkflowRunStatus(ctx context.Context, req UpdateWorkflowRunStatusRequest) error {
	res := a.db.WithContext(ctx).
		Model(&app.WorkflowRun{}).
		Where(app.WorkflowRun{ID: req.RunID}).
		Update("status", req.Status)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update workflow run status")
	}
	return nil
}
