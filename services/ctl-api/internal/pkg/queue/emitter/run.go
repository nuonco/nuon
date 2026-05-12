package emitter

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
)

func (e *emitterWorkflow) run(ctx workflow.Context) (finished bool, err error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	l.Info("registering emitter handlers")
	if err := e.registerHandlers(ctx); err != nil {
		return false, errors.Wrap(err, "unable to register handlers")
	}

	// Fetch the emitter from DB and set e.queueID from the authoritative record.
	emitter, err := e.ensureEmitterActive(ctx)
	if err != nil {
		return false, errors.Wrap(err, "unable to check emitter status")
	}
	if e.stopped {
		l.Info("emitter not found, stopping")
		return true, nil
	}

	// Check if the queue still exists before proceeding.
	if err := e.ensureQueueActive(ctx); err != nil {
		return false, errors.Wrap(err, "unable to check queue status")
	}
	if e.stopped {
		l.Info("queue terminated, stopping emitter")
		return true, nil
	}

	switch emitter.Mode {
	case app.QueueEmitterModeCron:
		return e.runCronMode(ctx, l, emitter)
	case app.QueueEmitterModeScheduled, app.QueueEmitterModeFireOnce:
		return e.runScheduledMode(ctx, l, emitter)
	default:
		return false, errors.Errorf("unknown emitter mode: %s", emitter.Mode)
	}
}

func (e *emitterWorkflow) emitLifecycleMetric(ctx workflow.Context, name string, emitter *app.QueueEmitter) {
	tags := metrics.ToTags(map[string]string{
		"signal_type": string(emitter.SignalType),
		"mode":        string(emitter.Mode),
		"owner_type":  emitter.Queue.OwnerType,
	})
	e.mw.Incr(ctx, name, tags...)
}

func (e *emitterWorkflow) emitSignalMetric(ctx workflow.Context, emitter *app.QueueEmitter, status string) {
	tags := metrics.ToTags(map[string]string{
		"signal_type":  string(emitter.SignalType),
		"emitter_type": string(emitter.Mode),
		"owner_type":   emitter.Queue.OwnerType,
		"status":       status,
	})
	e.mw.Incr(ctx, "queue.emitter.signal_emitted", tags...)
}

func (e *emitterWorkflow) emitSignal(ctx workflow.Context, l *zap.Logger, emitter *app.QueueEmitter) error {
	// Emit the signal to the queue and get back the signal ref
	resp, err := activities.AwaitEmitSignal(ctx, &activities.EmitSignalRequest{
		EmitterID: e.emitterID,
		QueueID:   emitter.QueueID,
	})
	if err != nil {
		e.emitSignalMetric(ctx, emitter, "error")
		return errors.Wrap(err, "unable to emit signal")
	}

	if resp.Skipped {
		e.emitSignalMetric(ctx, emitter, "skipped")
		l.Info("signal emission skipped - emitter already has in-flight signal",
			zap.String("emitter-id", e.emitterID),
			zap.String("queue-id", emitter.QueueID),
		)
		return nil
	}

	e.emitSignalMetric(ctx, emitter, "ok")

	l.Info("signal emitted, updating relationship",
		zap.String("queue-signal-id", resp.QueueSignalID),
		zap.String("workflow-id", resp.WorkflowID),
	)

	// Update the queue signal with the emitter relationship
	if _, err := activities.AwaitUpdateSignalEmitter(ctx, &activities.UpdateSignalEmitterRequest{
		QueueSignalID: resp.QueueSignalID,
		EmitterID:     e.emitterID,
	}); err != nil {
		return errors.Wrap(err, "unable to update signal emitter relationship")
	}

	// Update emitter stats
	if _, err := activities.AwaitUpdateEmitterStats(ctx, &activities.UpdateEmitterStatsRequest{
		EmitterID: e.emitterID,
	}); err != nil {
		l.Warn("failed to update emitter stats", zap.Error(err))
	}

	return nil
}
