package client

import (
	"context"
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// RetryStepRequest is the input for retrying a workflow step.
type RetryStepRequest struct {
	InstallWorkflowID string
	StepID            string
}

// RetryStepResponse is the response from the retry-step update.
type RetryStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Retryable  bool   `json:"retryable"`
}

// RetryStep sends a "retry-step" update to the execute-flow handler workflow
// for the given install workflow. Uses update-with-start so the handler is
// restarted if it has already terminated.
func (c *Client) RetryStep(ctx context.Context, req *RetryStepRequest) (*RetryStepResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	startOp := c.tClient.NewWithStartWorkflowOperation(
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

	_, err = c.tClient.UpdateWithStartWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWithStartWorkflowOptions{
			UpdateOptions: tclient.UpdateWorkflowOptions{
				WorkflowID:   qs.Workflow.ID,
				UpdateName:   "retry-step",
				WaitForStage: tclient.WorkflowUpdateStageAccepted,
				Args: []any{
					executeflow.RetryStepRequest{
						StepID: req.StepID,
					},
				},
			},
			StartWorkflowOperation: startOp,
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send retry-step update: %w", err)
	}

	return &RetryStepResponse{
		WorkflowID: qs.Workflow.ID,
		Retryable:  true,
	}, nil
}
