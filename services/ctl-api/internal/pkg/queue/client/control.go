package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) Pause(ctx context.Context, queueID string) error {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue")
	}

	update, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   q.Workflow.ID,
			UpdateName:   queue.PauseHandlerName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				queue.PauseRequest{},
			},
		},
		StartWorkflowOperation: c.queueStartOperation(q),
	})
	if err != nil {
		return errors.Wrap(err, "unable to call update handler")
	}

	var resp queue.PauseResponse
	if err := update.Get(ctx, &resp); err != nil {
		return errors.Wrap(err, "error waiting for handler to finish")
	}

	return nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) Resume(ctx context.Context, queueID string) error {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue")
	}

	update, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   q.Workflow.ID,
			UpdateName:   queue.ResumeHandlerName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				queue.ResumeRequest{},
			},
		},
		StartWorkflowOperation: c.queueStartOperation(q),
	})
	if err != nil {
		return errors.Wrap(err, "unable to call update handler")
	}

	var resp queue.ResumeResponse
	if err := update.Get(ctx, &resp); err != nil {
		return errors.Wrap(err, "error waiting for handler to finish")
	}

	return nil
}
