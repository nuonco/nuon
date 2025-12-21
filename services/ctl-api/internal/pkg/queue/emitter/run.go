package emitter

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
)

func (e *emitterWorkflow) run(ctx workflow.Context) (bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	l.Info("registering emitter handlers")
	if err := e.registerHandlers(ctx); err != nil {
		return false, errors.Wrap(err, "unable to register handlers")
	}

	l.Info("fetching emitter configuration")
	emitter, err := activities.AwaitGetEmitter(ctx, &activities.GetEmitterRequest{
		EmitterID: e.emitterID,
	})
	if err != nil {
		return false, errors.Wrap(err, "unable to get emitter")
	}

	switch emitter.Mode {
	case app.QueueEmitterModeCron:
		return e.runCronMode(ctx, l, emitter)
	case app.QueueEmitterModeScheduled:
		return e.runScheduledMode(ctx, l, emitter)
	default:
		return false, errors.Errorf("unknown emitter mode: %s", emitter.Mode)
	}
}

func (e *emitterWorkflow) emitSignal(ctx workflow.Context, l *zap.Logger, emitter *app.QueueEmitter) error {
	// Emit the signal to the queue and get back the signal ref
	resp, err := activities.AwaitEmitSignal(ctx, &activities.EmitSignalRequest{
		EmitterID: e.emitterID,
		QueueID:   emitter.QueueID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to emit signal")
	}

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
