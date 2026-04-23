package client

import (
	"context"

	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/callback"
)

// fireCallbacks loads and invokes all callbacks for the given signal and event.
// This is best-effort: errors are logged but do not fail the caller.
func (c *Client) fireCallbacks(ctx context.Context, queueSignalID string, queueID string, event callback.Event) {
	var callbacks []app.QueueSignalCallback
	res := c.db.WithContext(ctx).Where(app.QueueSignalCallback{
		QueueSignalID: queueSignalID,
		Event:         string(event),
	}).Find(&callbacks)
	if res.Error != nil {
		c.l.Warn("unable to query callbacks",
			zap.String("queue_signal_id", queueSignalID),
			zap.String("event", string(event)),
			zap.Error(res.Error))
		return
	}

	if len(callbacks) == 0 {
		return
	}

	payload := callback.CallbackPayload{
		Event:         event,
		QueueSignalID: queueSignalID,
		QueueID:       queueID,
	}

	for _, cb := range callbacks {
		uh := cb.UpdateHandler
		_, err := c.tClient.UpdateWorkflowInNamespace(ctx, uh.Namespace, tclient.UpdateWorkflowOptions{
			WorkflowID:   uh.WorkflowID,
			UpdateName:   uh.UpdateName,
			WaitForStage: tclient.WorkflowUpdateStageAccepted,
			Args:         []any{payload},
		})
		if err != nil {
			c.l.Warn("callback invocation failed",
				zap.String("callback_id", cb.ID),
				zap.String("workflow_id", uh.WorkflowID),
				zap.String("update_name", uh.UpdateName),
				zap.Error(err))
			continue
		}
	}
}
