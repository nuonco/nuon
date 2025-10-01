package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetFlowStepRequest struct {
	FlowStepID string `json:"flow_step_id" validate:"required"`
}

// @temporal-gen activity
// @by-id FlowStepID
func (a *Activities) PkgWorkflowsFlowGetFlowsStep(ctx context.Context, req GetFlowStepRequest) (*app.WorkflowStep, error) {
	var step app.WorkflowStep

	res := a.db.WithContext(ctx).
		Preload("Approval").
		Preload("Approval.Response").
		First(&step, "id = ?", req.FlowStepID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow step")
	}

	return &step, nil
}
