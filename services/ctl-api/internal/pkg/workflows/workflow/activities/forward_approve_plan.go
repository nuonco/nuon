package activities

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

type ForwardApprovePlanRequest struct {
	StepID             string `json:"step_id" validate:"required"`
	ApprovalResponseID string `json:"approval_response_id"`
	ResponseType       string `json:"response_type"`
}

type ForwardApprovePlanResponse struct {
	StepID string `json:"step_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardApprovePlan(ctx context.Context, req ForwardApprovePlanRequest) (*ForwardApprovePlanResponse, error) {
	// why: Find the step's handler workflow via the queue_signals table.
	// Filter by signal type to avoid matching the inner signal (same OwnerID/OwnerType).
	var qs app.QueueSignal
	res := a.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   req.StepID,
			OwnerType: (&app.WorkflowStep{}).TableName(),
			Type:      "execute-workflow-step",
		}).
		Order("created_at DESC").
		First(&qs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find step queue signal for step %s: %w", req.StepID, res.Error)
	}

	type approvePlanArg struct {
		ApprovalResponseID string `json:"approval_response_id"`
		ResponseType       string `json:"response_type"`
	}

	handle, err := handler.UpdateWithStart(ctx, a.tClient, &qs, handler.UpdateWithStartOptions{
		UpdateName:   "approve-plan",
		WaitForStage: tclient.WorkflowUpdateStageAccepted,
		Args: []any{
			approvePlanArg{
				ApprovalResponseID: req.ApprovalResponseID,
				ResponseType:       req.ResponseType,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send approve-plan update to step %s: %w", req.StepID, err)
	}

	var result error
	if err := handle.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("approve-plan update failed for step %s: %w", req.StepID, err)
	}

	return &ForwardApprovePlanResponse{StepID: req.StepID}, nil
}
