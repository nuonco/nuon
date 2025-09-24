package queue

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (q *queue) run(ctx workflow.Context) (bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
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

	l.Info("starting workers")
	if err := q.startWorkers(ctx); err != nil {
		return false, errors.Wrap(err, "unable to start workers")
	}

	q.ready = true

	if err := workflow.Await(ctx, func() bool {
		return generics.AnyTrue(q.stopped, q.restarted)
	}); err != nil {
		return false, err
	}

	if q.restarted {
		return false, nil
	}
	if q.stopped {
		return true, nil
	}

	return false, nil
}
