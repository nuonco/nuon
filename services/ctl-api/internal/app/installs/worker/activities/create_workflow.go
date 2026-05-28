package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const workflowIdempotencyKeyMetadataKey = "idempotency_key"

type CreateWorkflowRequest struct {
	InstallID    string            `validate:"required"`
	WorkflowType app.WorkflowType  `validate:"required"`
	Metadata     map[string]string `validate:"required"`
	PlanOnly     bool
	// IdempotencyKey, when set, makes CreateWorkflow a get-or-create: a prior
	// Workflow row with the same key for the same install is returned instead
	// of a fresh row being inserted. Use a value stable across activity retries
	// (e.g. workflow.GetInfo(ctx).WorkflowExecution.ID).
	IdempotencyKey string
}

// @temporal-gen-v2 activity
func (a *Activities) CreateWorkflow(ctx context.Context, req CreateWorkflowRequest) (*app.Workflow, error) {
	if req.IdempotencyKey != "" {
		var existing app.Workflow
		err := a.db.WithContext(ctx).
			Where(app.Workflow{
				OwnerID:   req.InstallID,
				OwnerType: "installs",
			}).
			Where("metadata -> ? = ?", workflowIdempotencyKeyMetadataKey, req.IdempotencyKey).
			First(&existing).Error
		if err == nil {
			return &existing, nil
		}
	}

	metadata := req.Metadata
	if req.IdempotencyKey != "" {
		if metadata == nil {
			metadata = map[string]string{}
		}
		metadata[workflowIdempotencyKeyMetadataKey] = req.IdempotencyKey
	}

	workflow, err := a.helpers.CreateWorkflow(ctx, req.InstallID, req.WorkflowType, metadata, req.PlanOnly)
	if err != nil {
		return nil, fmt.Errorf("unable to create workflow: %w", err)
	}

	return workflow, nil
}
