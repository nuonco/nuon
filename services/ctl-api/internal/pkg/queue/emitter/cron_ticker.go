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

const cronTickSecondJitterWindow = 45

type CronTickerWorkflowRequest struct {
	QueueID   string `validate:"required"`
	EmitterID string `validate:"required"`
}

func cronTickSecondJitter(workflowID string) time.Duration {
	h := fnv.New32a()
	_, _ = h.Write([]byte(workflowID))
	return time.Duration(h.Sum32()%cronTickSecondJitterWindow) * time.Second
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

	if offset := cronTickSecondJitter(workflow.GetInfo(ctx).WorkflowExecution.ID); offset > 0 {
		if err := workflow.Sleep(ctx, offset); err != nil {
			return err
		}
	}

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

	if emitter.Status.Status == app.StatusDisabled {
		l.Info("emitter disabled, terminating cron ticker",
			zap.String("emitter-id", req.EmitterID),
			zap.String("disabled-reason", emitter.Status.StatusHumanDescription),
		)
		info := workflow.GetInfo(ctx)
		_ = activities.AwaitTerminateWorkflow(ctx, &activities.TerminateWorkflowRequest{
			WorkflowID: info.WorkflowExecution.ID,
			Namespace:  info.Namespace,
			Reason:     "emitter disabled",
		})
		return nil
	}

	ctx = repairWorkflowContext(ctx, emitter)

	if emitter.Status.Status == app.StatusCancelled {
		l.Info("emitter is paused, skipping emit")
		return nil
	}

	if w.cfg.DisableEmitterSignals {
		l.Info("emitter signals disabled globally, skipping emit",
			zap.String("queue-id", req.QueueID))
		return nil
	}

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

	if _, err := activities.AwaitUpdateSignalEmitter(ctx, &activities.UpdateSignalEmitterRequest{
		QueueSignalID: resp.QueueSignalID,
		EmitterID:     emitter.ID,
	}); err != nil {
		return err
	}

	if _, err := activities.AwaitUpdateEmitterStats(ctx, &activities.UpdateEmitterStatsRequest{
		EmitterID: emitter.ID,
	}); err != nil {
		l.Warn("failed to update emitter stats", zap.Error(err))
	}

	return nil
}
