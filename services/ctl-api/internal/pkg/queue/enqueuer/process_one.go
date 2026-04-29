package enqueuer

import (
	"context"
	"time"

	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

// processOne looks up the queue signal and its parent queue, performs the
// UpdateWithStart call, and marks the signal as enqueued.
func (e *Enqueuer) processOne(queueSignalID string) {
	ctx, cancel := context.WithTimeout(e.ctx, processOneTimeout)
	defer cancel()

	var qs app.QueueSignal
	if res := e.db.WithContext(ctx).First(&qs, "id = ?", queueSignalID); res.Error != nil {
		e.l.Warn("unable to get queue signal for enqueue",
			zap.String("queue-signal-id", queueSignalID),
			zap.Error(res.Error))
		return
	}

	var q app.Queue
	if res := e.db.WithContext(ctx).First(&q, "id = ?", qs.QueueID); res.Error != nil {
		e.l.Warn("unable to get queue for enqueue",
			zap.String("queue-signal-id", queueSignalID),
			zap.String("queue-id", qs.QueueID),
			zap.Error(res.Error))
		return
	}

	startOp := e.queueStartOperation(&q)

	_, err := e.tClient.UpdateWithStartWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   q.Workflow.ID,
			UpdateName:   queue.EnqueueUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageAccepted,
			Args: []any{
				queue.EnqueueHandlerInput{
					QueueSignalID: qs.ID,
					WorkflowID:    qs.Workflow.ID,
				},
			},
		},
		StartWorkflowOperation: startOp,
	})

	metadata := map[string]any{
		"enqueue_started_at":  time.Now().UTC().Format(time.RFC3339),
		"enqueue_finished_at": time.Now().UTC().Format(time.RFC3339),
	}
	if err != nil {
		metadata["enqueue_error"] = err.Error()
		e.l.Warn("background enqueue failed",
			zap.String("queue-signal-id", queueSignalID),
			zap.Error(err))
	} else {
		if res := e.db.WithContext(ctx).
			Model(&app.QueueSignal{}).
			Where("id = ?", queueSignalID).
			Update("enqueued", true); res.Error != nil {
			e.l.Warn("unable to mark signal as enqueued",
				zap.String("queue-signal-id", queueSignalID),
				zap.Error(res.Error))
		}
	}

	if mergeErr := generics.MergeJSONBMetadata(e.db.WithContext(ctx), &app.QueueSignal{}, queueSignalID, "status", metadata); mergeErr != nil {
		e.l.Warn("unable to update queue signal metadata",
			zap.String("queue-signal-id", queueSignalID),
			zap.Error(mergeErr))
	}
}
