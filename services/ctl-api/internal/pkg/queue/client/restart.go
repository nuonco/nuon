package client

import (
	"context"

	"github.com/pkg/errors"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) Restart(ctx context.Context, queueID string) error {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue")
	}

	// Create workflow start operation for update-with-start
	wkflowReq := queue.QueueWorkflowRequest{
		QueueID: q.ID,
		Version: c.cfg.Version,
	}
	startOpts := tclient.StartWorkflowOptions{
		ID:        q.Workflow.ID,
		TaskQueue: workflows.APITaskQueue,
		Memo: map[string]any{
			"id":         q.ID,
			"owner-id":   q.OwnerID,
			"owner-type": q.OwnerType,
		},
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}

	// Build the update-with-start options
	update, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   q.Workflow.ID,
			UpdateName:   queue.RestartUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				queue.RestartRequest{},
			},
		},
		StartWorkflowOperation: c.tClient.NewWithStartWorkflowOperation(startOpts, "Queue", wkflowReq),
	})
	if err != nil {
		return errors.Wrap(err, "unable to call update handler")
	}

	var resp queue.RestartResponse
	if err := update.Get(ctx, &resp); err != nil {
		return errors.Wrap(err, "error waiting for handler to finish")
	}

	return nil
}
