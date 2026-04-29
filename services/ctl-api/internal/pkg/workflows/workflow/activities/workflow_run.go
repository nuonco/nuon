package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
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

	// If the context doesn't have an account ID (e.g. running inside a queue
	// handler workflow), fall back to the parent workflow's CreatedByID so
	// the NOT NULL constraint is satisfied.
	if keys.CreatedByIDFromContext(ctx) == "" {
		var wf app.Workflow
		if res := a.db.WithContext(ctx).Select("created_by_id", "org_id").First(&wf, "id = ?", req.WorkflowID); res.Error == nil {
			run.CreatedByID = wf.CreatedByID
			run.OrgID = wf.OrgID
		}
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
