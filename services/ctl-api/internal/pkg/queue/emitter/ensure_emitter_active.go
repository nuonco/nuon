package emitter

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
)

func (e *emitterWorkflow) ensureEmitterActive(ctx workflow.Context) (*app.QueueEmitter, error) {
	l, _ := log.WorkflowLogger(ctx)

	emitter, err := activities.AwaitGetEmitterByEmitterID(ctx, e.emitterID)
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			l.Warn("emitter not found, stopping workflow", zap.String("emitter-id", e.emitterID))
			e.stopped = true
			return nil, nil
		}
		return nil, err
	}

	if emitter.Status.Status == app.StatusDisabled {
		l.Info("emitter disabled, stopping workflow",
			zap.String("emitter-id", e.emitterID),
			zap.String("disabled-reason", emitter.Status.StatusHumanDescription),
		)
		e.stopped = true
		return nil, nil
	}

	// Set the queue ID from the database record. This is the source of truth
	// and fixes a race where the workflow request's QueueID might be empty.
	e.queueID = emitter.QueueID

	return emitter, nil
}
