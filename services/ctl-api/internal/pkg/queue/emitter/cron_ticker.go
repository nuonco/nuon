package emitter

import (
	"hash/fnv"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
)

// jitterOffset returns a deterministic offset in [0, window) for the given
// emitter ID, used to spread cron emitter ticks across a configurable window.
func jitterOffset(emitterID string, window time.Duration) time.Duration {
	if window <= 0 {
		return 0
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(emitterID))
	return time.Duration(uint64(h.Sum32()) % uint64(window))
}

type CronTickerWorkflowRequest struct {
	QueueID   string `validate:"required"`
	EmitterID string `validate:"required"`
}

// @temporal-gen-v2 workflow
// @task-queue "queue"
// @id-template queue-emitter-cron-{{.QueueID}}-{{.EmitterID}}
func (w *Workflows) CronTicker(ctx workflow.Context, req CronTickerWorkflowRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("cron ticker fired",
		zap.String("emitter-id", req.EmitterID),
		zap.String("queue-id", req.QueueID),
	)

	// Fetch emitter to check status and get signal template
	emitter, err := activities.AwaitGetEmitter(ctx, &activities.GetEmitterRequest{
		EmitterID: req.EmitterID,
	})
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			l.Warn("emitter not found, terminating orphaned cron ticker",
				zap.String("emitter-id", req.EmitterID),
			)
			info := workflow.GetInfo(ctx)
			_ = activities.AwaitTerminateWorkflow(ctx, &activities.TerminateWorkflowRequest{
				WorkflowID: info.WorkflowExecution.ID,
				Namespace:  info.Namespace,
				Reason:     "emitter not found",
			})
			return nil
		}
		l.Error("failed to get emitter", zap.Error(err))
		return err
	}

	// Check if emitter is paused (status is cancelled)
	if emitter.Status.Status == app.StatusCancelled {
		l.Info("emitter is paused, skipping emit")
		return nil
	}

	if offset := jitterOffset(emitter.ID, emitter.JitterWindow); offset > 0 {
		if err := workflow.Sleep(ctx, offset); err != nil {
			return err
		}
	}

	// Emit the signal
	if err := w.emitSignal(ctx, l, emitter); err != nil {
		l.Error("failed to emit signal", zap.Error(err))
		return err
	}

	l.Info("emit complete")
	return nil
}

func (w *Workflows) emitSignalMetric(ctx workflow.Context, emitter *app.QueueEmitter, status string) {
	tags := metrics.ToTags(map[string]string{
		"signal_type":  string(emitter.SignalType),
		"emitter_type": string(emitter.Mode),
		"owner_type":   emitter.Queue.OwnerType,
		"status":       status,
	})
	w.mw.Incr(ctx, "queue.emitter.signal_emitted", tags...)
}

func (w *Workflows) emitSignal(ctx workflow.Context, l *zap.Logger, emitter *app.QueueEmitter) error {
	// Emit the signal to the queue and get back the signal ref
	resp, err := activities.AwaitEmitSignal(ctx, &activities.EmitSignalRequest{
		EmitterID: emitter.ID,
		QueueID:   emitter.QueueID,
	})
	if err != nil {
		w.emitSignalMetric(ctx, emitter, "error")
		return err
	}

	if resp.Skipped {
		w.emitSignalMetric(ctx, emitter, "skipped")
		l.Info("signal emission skipped - emitter already has in-flight signal",
			zap.String("emitter-id", emitter.ID),
			zap.String("queue-id", emitter.QueueID),
		)
		return nil
	}

	w.emitSignalMetric(ctx, emitter, "ok")

	l.Info("signal emitted, updating relationship",
		zap.String("queue-signal-id", resp.QueueSignalID),
		zap.String("workflow-id", resp.WorkflowID),
	)

	// Update the queue signal with the emitter relationship
	if _, err := activities.AwaitUpdateSignalEmitter(ctx, &activities.UpdateSignalEmitterRequest{
		QueueSignalID: resp.QueueSignalID,
		EmitterID:     emitter.ID,
	}); err != nil {
		return err
	}

	// Update emitter stats
	if _, err := activities.AwaitUpdateEmitterStats(ctx, &activities.UpdateEmitterStatsRequest{
		EmitterID: emitter.ID,
	}); err != nil {
		l.Warn("failed to update emitter stats", zap.Error(err))
	}

	return nil
}
