package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// CheckWorkflowRetryableRequest is the input for checking if a workflow is retryable.
type CheckWorkflowRetryableRequest struct {
	WorkflowID string `json:"workflow_id" validate:"required"`
}

// CheckWorkflowRetryableResponse is the output indicating retryability.
type CheckWorkflowRetryableResponse struct {
	Retryable bool `json:"retryable"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) CheckWorkflowRetryable(ctx context.Context, req CheckWorkflowRetryableRequest) (*CheckWorkflowRetryableResponse, error) {
	var workflow app.Workflow
	if res := a.db.WithContext(ctx).First(&workflow, "id = ?", req.WorkflowID); res.Error != nil {
		return nil, fmt.Errorf("unable to get workflow: %w", res.Error)
	}

	// A workflow is not retryable if it succeeded
	if workflow.Status.Status == app.StatusSuccess {
		return &CheckWorkflowRetryableResponse{Retryable: false}, nil
	}

	// A workflow is not retryable if a newer workflow for the same owner has been started
	var newerCount int64
	res := a.db.WithContext(ctx).Model(&app.Workflow{}).
		Where("owner_id = ? AND owner_type = ? AND id != ? AND created_at > ?",
			workflow.OwnerID, workflow.OwnerType, workflow.ID, workflow.CreatedAt).
		Count(&newerCount)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to check newer workflows: %w", res.Error)
	}
	if newerCount > 0 {
		return &CheckWorkflowRetryableResponse{Retryable: false}, nil
	}

	return &CheckWorkflowRetryableResponse{Retryable: true}, nil
}
