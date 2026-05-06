package enqueuer

import (
	"context"
	"time"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuecctx"
)

// EnqueueSource identifies how a signal was enqueued.
const (
	EnqueueSourceChannel = "channel"
	EnqueueSourceAwait   = "await"
	EnqueueSourceSweep   = "sweep"
)

// EnqueueInline synchronously enqueues a queue signal by performing the
// UpdateWithStart call inline with the caller. It records enqueue timing
// metadata (including the enqueue source) and marks the signal as enqueued
// on success.
func (e *Enqueuer) EnqueueInline(ctx context.Context, queueSignalID string, source string) error {
	var qs app.QueueSignal
	if res := e.db.WithContext(ctx).First(&qs, "id = ?", queueSignalID); res.Error != nil {
		return errors.Wrap(res.Error, "unable to get queue signal for enqueue")
	}

	if qs.Enqueued {
		return nil
	}

	var q app.Queue
	if res := e.db.WithContext(ctx).First(&q, "id = ?", qs.QueueID); res.Error != nil {
		return errors.Wrap(res.Error, "unable to get queue for enqueue")
	}

	ctx = queuecctx.Apply(ctx, qs.SignalContext)

	enqueueStartedAt := time.Now().UTC().Format(time.RFC3339)

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

	enqueueFinishedAt := time.Now().UTC().Format(time.RFC3339)

	metadata := map[string]any{
		"enqueue_started_at":  enqueueStartedAt,
		"enqueue_finished_at": enqueueFinishedAt,
		"enqueue_source":      source,
	}
	if err != nil {
		metadata["enqueue_error"] = err.Error()
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

	if err != nil {
		return errors.Wrap(err, "enqueue UpdateWithStart failed")
	}

	return nil
}

// processOne looks up the queue signal and its parent queue, performs the
// UpdateWithStart call, and marks the signal as enqueued.
func (e *Enqueuer) processOne(queueSignalID string) {
	ctx, cancel := context.WithTimeout(e.ctx, processOneTimeout)
	defer cancel()

	if err := e.EnqueueInline(ctx, queueSignalID, EnqueueSourceChannel); err != nil {
		e.l.Warn("background enqueue failed",
			zap.String("queue-signal-id", queueSignalID),
			zap.Error(err))
	}
}
