package emitter

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
)

type CronTickerWorkflowRequest struct {
	QueueID   string `validate:"required"`
	EmitterID string `validate:"required"`
}

// @temporal-gen workflow
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
		l.Error("failed to get emitter", zap.Error(err))
		return err
	}

	// Check if emitter is paused (status is cancelled)
	if emitter.Status.Status == app.StatusCancelled {
		l.Info("emitter is paused, skipping emit")
		return nil
	}

	// Emit the signal
	if err := w.emitSignal(ctx, l, emitter); err != nil {
		l.Error("failed to emit signal", zap.Error(err))
		return err
	}

	l.Info("emit complete")
	return nil
}

func (w *Workflows) emitSignal(ctx workflow.Context, l *zap.Logger, emitter *app.QueueEmitter) error {
	// Emit the signal to the queue and get back the signal ref
	resp, err := activities.AwaitEmitSignal(ctx, &activities.EmitSignalRequest{
		EmitterID: emitter.ID,
		QueueID:   emitter.QueueID,
	})
	if err != nil {
		return err
	}

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
