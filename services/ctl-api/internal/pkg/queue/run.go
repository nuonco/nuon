package queue

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (q *queue) run(ctx workflow.Context) (bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	l.Info("ensuring queue is active")
	if err := q.ensureActive(ctx); err != nil {
		return false, errors.Wrap(err, "unable to ensure queue is active")
	}
	if q.stopped {
		return true, nil
	}

	l.Info("registering handlers")
	if err := q.registerHandlers(ctx); err != nil {
		return false, errors.Wrap(err, "unable to register handlers")
	}

	l.Info("setting up queue channels")
	if err := q.setupChannels(ctx); err != nil {
		return false, errors.Wrap(err, "unable to setup channels")
	}

	l.Info("requeuing any remaining signals")
	requeued, err := q.requeueSignals(ctx)
	if err != nil {
		return false, errors.Wrap(err, "unable to requeue signals")
	}

	// Restore lastActivityTime from state (survives continue-as-new),
	// or initialize for the first run.
	// If we just requeued pending signals, treat that as activity so we dont exit immediately using inherited
	// from previous run.
	switch {
	case requeued > 0:
		q.lastActivityTime = workflow.Now(ctx)
	case !q.state.LastActivityTime.IsZero():
		q.lastActivityTime = q.state.LastActivityTime
	default:
		q.lastActivityTime = workflow.Now(ctx)
	}

	l.Info("starting workers")
	if err := q.startWorkers(ctx); err != nil {
		return false, errors.Wrap(err, "unable to start workers")
	}

	q.ready = true

	for {
		if _, err := workflow.AwaitWithTimeout(ctx, queueReceiveTimeout, func() bool {
			return generics.AnyTrue(q.stopped, q.restarted, q.shouldContinueAsNew(ctx)) ||
				(q.isIdle(ctx) && q.activeWorkers == 0)
		}); err != nil {
			return false, err
		}

		if q.restarted {
			return false, nil
		}
		if q.stopped {
			return true, nil
		}
		if q.isIdle(ctx) && q.activeWorkers == 0 {
			l.Info("queue is idle, terminating workflow")
			return true, nil
		}

		if q.shouldContinueAsNew(ctx) {
			l.Info("workflow history growing large, continuing as new",
				zap.Int("history_length", workflow.GetInfo(ctx).GetCurrentHistoryLength()),
				zap.Bool("server_suggested", workflow.GetInfo(ctx).GetContinueAsNewSuggested()),
			)
			return false, nil
		}
	}
}
