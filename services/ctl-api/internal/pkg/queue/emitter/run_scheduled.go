package emitter

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
)

func (e *emitterWorkflow) runScheduledMode(ctx workflow.Context, l *zap.Logger, emitter *app.QueueEmitter) (bool, error) {
	l.Info("running in scheduled mode",
		zap.String("emitter-id", e.emitterID),
		zap.String("queue-id", emitter.QueueID),
		zap.Timep("scheduled-at", emitter.ScheduledAt),
	)

	// Check if already fired
	if emitter.Fired {
		l.Info("scheduled emitter already fired, stopping")
		return true, nil
	}

	if emitter.ScheduledAt == nil {
		return false, errors.New("scheduled emitter has no scheduled_at time")
	}

	// Calculate how long to wait
	now := workflow.Now(ctx)
	waitDuration := emitter.ScheduledAt.Sub(now)

	if waitDuration > 0 {
		l.Info("waiting until scheduled time",
			zap.Duration("wait-duration", waitDuration),
			zap.Time("scheduled-at", *emitter.ScheduledAt),
		)

		// Wait until the scheduled time, but also listen for stop signal
		timerFuture := workflow.NewTimer(ctx, waitDuration)

		selector := workflow.NewSelector(ctx)
		var timerFired bool

		selector.AddFuture(timerFuture, func(f workflow.Future) {
			timerFired = true
		})

		selector.Select(ctx)

		// Check if we were stopped while waiting
		if e.stopped {
			l.Info("emitter stopped while waiting")
			return true, nil
		}

		if !timerFired {
			// Continue-as-new to refresh state and try again
			return false, nil
		}
	}

	// Fire the signal
	l.Info("scheduled time reached, emitting signal")

	if err := e.emitSignal(ctx, l, emitter); err != nil {
		return false, err
	}

	// Mark as fired in the database
	if _, err := activities.AwaitMarkEmitterFired(ctx, &activities.MarkEmitterFiredRequest{
		EmitterID: e.emitterID,
	}); err != nil {
		l.Warn("failed to mark emitter as fired", zap.Error(err))
	}

	e.state.EmitCount++

	l.Info("scheduled emit complete, stopping emitter",
		zap.Int64("total-emit-count", e.state.EmitCount),
	)

	// Return true to indicate workflow is finished (no continue-as-new)
	return true, nil
}
