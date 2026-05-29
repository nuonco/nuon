package emitter

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
)

type ScheduleEmitRequest struct {
	EmitterID string `validate:"required"`
	QueueID   string `validate:"required"`
}

// ScheduleEmit is the action target for Temporal-native Schedules. Each scheduled
// fire runs this short-lived workflow, which emits the emitter's signal onto its
// queue. It is the native-scheduling counterpart to CronTicker and reuses the same
// emitSignal path (and therefore the same in-flight dedup + queue serialization).
//
// @temporal-gen-v2 workflow
// @task-queue "queue"
// @id-template queue-schedule-emit-{{.QueueID}}-{{.EmitterID}}
func (w *Workflows) ScheduleEmit(ctx workflow.Context, req ScheduleEmitRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	// Work-layer gate: when native scheduling is disabled the legacy CronTicker is
	// authoritative, so a leftover schedule must not double-emit. No-op.
	if !w.cfg.NativeSchedulingEnabled {
		l.Info("native scheduling disabled, schedule-emit no-op", zap.String("emitter-id", req.EmitterID))
		return nil
	}

	em, err := activities.AwaitGetEmitter(ctx, &activities.GetEmitterRequest{EmitterID: req.EmitterID})
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			l.Warn("emitter not found, schedule-emit no-op", zap.String("emitter-id", req.EmitterID))
			return nil
		}
		return err
	}

	// Paused emitters skip emission (status set to cancelled).
	if em.Status.Status == app.StatusCancelled {
		l.Info("emitter is paused, skipping emit")
		return nil
	}

	return w.emitSignal(ctx, l, em)
}
