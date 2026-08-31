package enqueuer

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
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

var errOrphanedQueueSignal = errors.New("queue signal references a missing queue")

// EnqueueInline synchronously enqueues a queue signal by performing the
// SignalWithStart call inline with the caller. It records enqueue timing
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
		if stderrors.Is(res.Error, gorm.ErrRecordNotFound) {
			return e.quarantineOrphanedSignal(ctx, &qs)
		}
		return errors.Wrap(res.Error, "unable to get queue for enqueue")
	}

	ctx = queuecctx.Apply(ctx, qs.SignalContext)

	enqueueStart := time.Now().UTC()
	enqueueStartedAt := enqueueStart.Format(time.RFC3339)

	startOpts, wkflowReq := e.queueStartOptions(&q)
	_, err := e.tClient.SignalWithStartWorkflowInNamespace(ctx, q.Workflow.Namespace,
		q.Workflow.ID,
		queue.EnqueueSignalName,
		queue.EnqueueHandlerInput{
			QueueSignalID: qs.ID,
			WorkflowID:    qs.Workflow.ID,
		},
		startOpts,
		"Queue",
		wkflowReq,
	)

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
		failTags := metrics.ToTags(map[string]string{
			"source":           source,
			"success":          "false",
			"signal_namespace": q.Workflow.Namespace,
		})
		e.mw.Incr("queue_signals.enqueue", failTags)
		e.mw.Timing("queue_signals.enqueue.latency", time.Since(enqueueStart), failTags)
		return errors.Wrap(err, "enqueue SignalWithStart failed")
	}

	enqueueTags := metrics.ToTags(map[string]string{
		"source":           source,
		"success":          "true",
		"signal_namespace": q.Workflow.Namespace,
	})
	e.mw.Incr("queue_signals.enqueue", enqueueTags)
	e.mw.Timing("queue_signals.enqueue.latency", time.Since(enqueueStart), enqueueTags)

	return nil
}

func (e *Enqueuer) quarantineOrphanedSignal(ctx context.Context, qs *app.QueueSignal) error {
	status := app.NewCompositeStatus(ctx, app.StatusError)
	status.StatusHumanDescription = "parent queue not found; signal quarantined before enqueue"
	status.Metadata = map[string]any{
		"queue_id": qs.QueueID,
		"reason":   "missing_parent_queue",
	}

	res := e.db.WithContext(ctx).
		Model(&app.QueueSignal{}).
		Where(&app.QueueSignal{ID: qs.ID}).
		Where(map[string]any{"deleted_at": 0, "enqueued": false}).
		Updates(map[string]any{
			"deleted_at": time.Now().Unix(),
			"status":     &status,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to quarantine queue signal with missing queue")
	}

	if res.RowsAffected > 0 {
		tags := metrics.ToTags(map[string]string{
			"signal_type": string(qs.Type),
			"owner_type":  qs.OwnerType,
		})
		e.mw.Incr("queue_signals.enqueue.orphaned", tags)
		e.l.Warn("quarantined queue signal with missing parent queue",
			zap.String("queue-signal-id", qs.ID),
			zap.String("queue-id", qs.QueueID),
			zap.String("signal-type", string(qs.Type)),
			zap.String("owner-id", qs.OwnerID),
			zap.String("owner-type", qs.OwnerType))
	}

	return errOrphanedQueueSignal
}

// processOne looks up the queue signal and its parent queue, performs the
// SignalWithStart call, and marks the signal as enqueued.
func (e *Enqueuer) processOne(queueSignalID string) {
	ctx, cancel := context.WithTimeout(e.ctx, processOneTimeout)
	defer cancel()

	if err := e.EnqueueInline(ctx, queueSignalID, EnqueueSourceChannel); err != nil {
		if stderrors.Is(err, errOrphanedQueueSignal) {
			return
		}
		e.l.Warn("background enqueue failed",
			zap.String("queue-signal-id", queueSignalID),
			zap.Error(err))
	}
}
