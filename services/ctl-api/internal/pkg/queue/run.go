package queue

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

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
	if err := q.requeueSignals(ctx); err != nil {
		return false, errors.Wrap(err, "unable to requeue signals")
	}

	// Restore lastActivityTime from state (survives continue-as-new),
	// or initialize for the first run.
	if !q.state.LastActivityTime.IsZero() {
		q.lastActivityTime = q.state.LastActivityTime
	} else {
		q.lastActivityTime = workflow.Now(ctx)
	}

	l.Info("starting workers")
	if err := q.startWorkers(ctx); err != nil {
		return false, errors.Wrap(err, "unable to start workers")
	}

	q.ready = true

	if _, err := workflow.AwaitWithTimeout(ctx, queueReceiveTimeout, func() bool {
		return generics.AnyTrue(q.stopped, q.restarted) || (q.isIdle(ctx) && q.activeWorkers == 0)
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

	return false, nil
}
