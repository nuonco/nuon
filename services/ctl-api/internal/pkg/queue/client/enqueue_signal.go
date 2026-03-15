package client

import (
	"context"

	"github.com/pkg/errors"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type EnqueueSignalRequest struct {
	QueueID string        `validate:"required"`
	Signal  signal.Signal `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) EnqueueSignal(ctx context.Context, req *EnqueueSignalRequest) (*queue.EnqueueResponse, error) {
	q, err := c.getQueue(ctx, req.QueueID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	// Use UpdateWithStart so the queue workflow is automatically started if it is not
	// currently running (e.g. after a graceful stop).
	rawResp, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   q.Workflow.ID,
			UpdateName:   queue.EnqueueUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args:         []any{req.Signal},
		},
		StartWorkflowOperation: c.tClient.NewWithStartWorkflowOperation(
			tclient.StartWorkflowOptions{
				ID:        q.Workflow.ID,
				TaskQueue: workflows.APITaskQueue,
				Memo: map[string]any{
					"id":         q.ID,
					"owner-id":   q.OwnerID,
					"owner-type": q.OwnerType,
				},
				WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
				RetryPolicy:              &temporal.RetryPolicy{MaximumAttempts: 0},
			},
			"Queue",
			queue.QueueWorkflowRequest{QueueID: q.ID, Version: c.cfg.Version},
		),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call enqueue handler")
	}

	var resp queue.EnqueueResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable get response")
	}

	return &resp, nil
}
