package activities

import (
	"context"
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// ForwardCreateStepRetryRequest is the input for forwarding a create-step-retry to a step handler workflow.
type ForwardCreateStepRetryRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

// ForwardCreateStepRetryResponse is the output from forwarding a create-step-retry.
type ForwardCreateStepRetryResponse struct {
	StepID    string `json:"step_id"`
	NewStepID string `json:"new_step_id"`
	Directive string `json:"directive"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardCreateStepRetry(ctx context.Context, req ForwardCreateStepRetryRequest) (*ForwardCreateStepRetryResponse, error) {
	// Find the step's handler workflow via the queue_signals table.
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

	// Use update-with-start to ensure the handler workflow is alive.
	startOp := a.tClient.NewWithStartWorkflowOperation(
		tclient.StartWorkflowOptions{
			ID:                       qs.Workflow.ID,
			TaskQueue:                "api",
			WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 0,
			},
		},
		"Handler",
		handler.HandlerRequest{
			QueueID:       qs.QueueID,
			QueueSignalID: qs.ID,
		},
	)

	rawResp, err := a.tClient.UpdateWithStartWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWithStartWorkflowOptions{
			UpdateOptions: tclient.UpdateWorkflowOptions{
				WorkflowID:   qs.Workflow.ID,
				UpdateName:   "create-step-retry",
				WaitForStage: tclient.WorkflowUpdateStageCompleted,
			},
			StartWorkflowOperation: startOp,
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send create-step-retry update to step %s: %w", req.StepID, err)
	}

	type stepRetryResult struct {
		Directive string `json:"directive"`
		NewStepID string `json:"new_step_id"`
	}
	var result stepRetryResult
	if err := rawResp.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("create-step-retry update failed for step %s: %w", req.StepID, err)
	}

	return &ForwardCreateStepRetryResponse{
		StepID:    req.StepID,
		NewStepID: result.NewStepID,
		Directive: result.Directive,
	}, nil
}
