package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) CancelSignal(ctx context.Context, queueSignalID string) (*handler.CancelResponse, error) {
	q, err := c.getQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue signal")
	}

	// if already terminal, nothing to cancel
	if isTerminalStatus(q.Status.Status) {
		return &handler.CancelResponse{}, nil
	}

	rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   q.Workflow.ID,
		UpdateName:   handler.CancelUpdateName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			&handler.CancelRequest{},
		},
	})
	if err != nil {
		// workflow may be sleeping/completed — update DB directly
		c.updateQueueSignalStatus(ctx, queueSignalID, app.StatusCancelled)
		return &handler.CancelResponse{}, nil
	}

	var resp handler.CancelResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable to get response")
	}

	return &resp, nil
}

func (c *Client) updateQueueSignalStatus(ctx context.Context, queueSignalID string, status app.Status) {
	res := c.db.WithContext(ctx).
		Model(&app.QueueSignal{}).
		Where("id = ?", queueSignalID).
		Update("status", app.NewCompositeStatus(ctx, status))
	if res.Error != nil {
		c.l.Warn("failed to update queue signal status directly",
			zap.String("queue_signal_id", queueSignalID),
			zap.Error(res.Error),
		)
	}
}
