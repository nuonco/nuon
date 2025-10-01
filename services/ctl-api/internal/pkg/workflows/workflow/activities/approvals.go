package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateStepApprovalResponseRequest struct {
	StepApprovalID string                       `json:"step_approval_id" validate:"required"`
	Type           app.WorkflowStepResponseType `json:"type" validate:"required"`
	Note           string                       `json:"note" validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 1m
// @start-to-close-timeout 10s
func (a *Activities) CreateApprovalResponse(ctx context.Context, req CreateStepApprovalResponseRequest) (*app.WorkflowStepApprovalResponse, error) {
	approval := app.WorkflowStepApproval{}
	res := a.db.WithContext(ctx).Where(app.WorkflowStepApproval{ID: req.StepApprovalID}).First(&approval)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find approval with ID %s: %w", req.StepApprovalID, res.Error)
	}

	approvalResponse := app.WorkflowStepApprovalResponse{
		OrgID:                         approval.OrgID,
		CreatedBy:                     approval.CreatedBy,
		InstallWorkflowStepApprovalID: approval.ID,
		Type:                          req.Type,
		Note:                          req.Note,
	}
	res = a.db.WithContext(ctx).Create(&approvalResponse)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create approval response: %w", res.Error)
	}

	return &approvalResponse, nil
}
