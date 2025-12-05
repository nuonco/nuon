package temporal

import (
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"

	"github.com/cockroachdb/errors"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/pkg/shortid"
	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"

	actionssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/signals"
	appssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	componentssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	generalsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	installssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	orgssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	releasessignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/signals"
	runnerssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	signalsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/signals/activities"
)

const (
	defaultSignalSendTimeout time.Duration = time.Second * 5
)

func (e *evClient) Send(ctx workflow.Context, id string, signal eventloop.Signal) {
	l := temporalzap.GetWorkflowLogger(ctx)
	_, err := e.bootstrap(ctx, id, signal, l)
	if err != nil {
		return
	}

	switch signal.Namespace() {
	case actionssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendActionsSignal(ctx, &signalsactivities.SendSignalRequest[*actionssignals.Signal]{
			ID:     id,
			Signal: signal.(*actionssignals.Signal),
		})
	case appssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendAppsSignal(ctx, &signalsactivities.SendSignalRequest[*appssignals.Signal]{
			ID:     id,
			Signal: signal.(*appssignals.Signal),
		})
	case installssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendInstallsSignal(ctx, &signalsactivities.SendSignalRequest[*installssignals.Signal]{
			ID:     id,
			Signal: signal.(*installssignals.Signal),
		})
	case componentssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendComponentsSignal(ctx, &signalsactivities.SendSignalRequest[*componentssignals.Signal]{
			ID:     id,
			Signal: signal.(*componentssignals.Signal),
		})
	case orgssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendOrgsSignal(ctx, &signalsactivities.SendSignalRequest[*orgssignals.Signal]{
			ID:     id,
			Signal: signal.(*orgssignals.Signal),
		})
	case releasessignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendReleasesSignal(ctx, &signalsactivities.SendSignalRequest[*releasessignals.Signal]{
			ID:     id,
			Signal: signal.(*releasessignals.Signal),
		})
	case runnerssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendRunnersSignal(ctx, &signalsactivities.SendSignalRequest[*runnerssignals.Signal]{
			ID:     id,
			Signal: signal.(*runnerssignals.Signal),
		})
	case generalsignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendGeneralSignal(ctx, &signalsactivities.SendSignalRequest[*generalsignals.Signal]{
			ID:     id,
			Signal: signal.(*generalsignals.Signal),
		})
	default:
		err = errors.New("unsupported namespace " + signal.Namespace())
	}

	if err != nil {
		e.l.Error("unable to send signal",
			zap.String("id", id),
			zap.String("namespace", signal.Namespace()),
			zap.Error(err),
		)
	}
}

func (e *evClient) bootstrap(ctx workflow.Context, objectID string, signal eventloop.Signal, l *zap.Logger) (string, error) {
	if err := e.v.Struct(signal); err != nil {
		e.mw.Incr("event_loop.signal", metrics.ToStatusTag("invalid signal"))
		l.Error("invalid signal", zap.Error(err))
		return "", errors.Wrap(err, "invalid signal")
	}

	if err := signal.PropagateContext(ctx); err != nil {
		e.mw.Incr("event_loop.signal", metrics.ToStatusTag("unable to propagate"))
		l.Error("unable to propagate", zap.Error(err))
		return "", errors.Wrap(err, "unable to propagate values from context")
	}

	// This is a nasty hack to compensate for the fact that eventloop assumes an alignment between
	// workflow id and the signal channel to listen on. that's fine historically, but doesn't work when we have
	// subloops with a hierarchical id. The simplest solution is probably to not have hierarchical ids, but
	// for now we munge the id to include only the tail id, which should be the underlying object id that's being
	// listened on
	mungeid := objectID
	if idx := strings.LastIndex(objectID, "-"); idx != -1 {
		mungeid = objectID[idx+1:]
	}

	if signal.Start() || signal.Restart() {
		// For start and restart signals, ensure the target event loop is running

		// TODO(sdboyer) commented until we get this restored
		// if err := signalsactivities.AwaitStartEventLoop(ctx, &signalsactivities.StartEventLoopRequest{
		// 	WorkflowID:   "event-loop-" + objectID,
		// 	ObjectID:     mungeid,
		// 	Namespace:    signal.Namespace(),
		// 	WorkflowType: signal.WorkflowName(),
		// }); err != nil {
		// 	e.mw.Incr("event_loop.signal", metrics.ToStatusTag("unable_to_start_event_loop"))
		// 	return "", errors.Wrapf(err, "unable to start event loop for workflow with id %q in namespace %q", objectID, signal.Namespace())
		// }
	}

	return mungeid, nil
}

func (e *evClient) SendFaF(ctx workflow.Context, objectID string, signal eventloop.Signal) error {
	l := temporalzap.GetWorkflowLogger(ctx)
	mungeid, err := e.bootstrap(ctx, objectID, signal, l)
	if err != nil {
		return err
	}

	err = workflow.SignalExternalWorkflow(
		workflow.WithWorkflowNamespace(ctx, signal.Namespace()),
		signal.WorkflowID(objectID),
		"",
		mungeid,
		signal,
	).Get(ctx, nil)
	if err == nil {
		return nil
	}

	e.mw.Incr("event_loop.signal", metrics.ToStatusTag("unable_to_send"))
	l.Warn("unable to dispatch signal to workflow",
		zap.String("from-workflow", workflow.GetInfo(ctx).WorkflowExecution.ID),
		zap.String("to-workflow", objectID),
		zap.Error(err),
	)

	return errors.Wrapf(err, "failed sending signal to workflow with id %q in namespace %q", signal.WorkflowID(objectID), signal.Namespace())
}

func (e *evClient) SendAsync(ctx workflow.Context, objectID string, signal eventloop.Signal) (workflow.Future, error) {
	l := temporalzap.GetWorkflowLogger(ctx)
	mungeid, err := e.bootstrap(ctx, objectID, signal, l)
	if err != nil {
		return nil, err
	}

	var listenerID string
	if err := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return shortid.NewNanoID("listen")
	}).Get(&listenerID); err != nil {
		return nil, errors.Wrap(err, "unable to generate listener id")
	}
	if listenerID == "" {
		return nil, errors.New("generated listener id was empty")
	}

	info := workflow.GetInfo(ctx)
	listener := eventloop.SignalListener{
		WorkflowID: info.WorkflowExecution.ID,
		Namespace:  info.Namespace,
		SignalName: listenerID,
	}

	if err := eventloop.AppendListenerIDs(signal, listener); err != nil {
		e.mw.Incr("event_loop.signal", metrics.ToStatusTag("unable to add listeners"))
		l.Error("unable to register signal listeners", zap.Error(err))
		return nil, errors.Wrap(err, "unable to add listeners")
	}

	fut, set := workflow.NewFuture(ctx)

	err = workflow.SignalExternalWorkflow(
		workflow.WithWorkflowNamespace(ctx, signal.Namespace()),
		signal.WorkflowID(objectID),
		"",
		mungeid,
		signal,
	).Get(ctx, nil)
	if err != nil {
		e.mw.Incr("event_loop.signal", metrics.ToStatusTag("unable_to_send"))
		l.Warn("unable to dispatch signal to workflow",
			zap.String("from-workflow", info.WorkflowExecution.ID),
			zap.String("to-workflow", objectID),
			zap.Error(err),
		)

		return nil, errors.Wrapf(err, "failed sending signal to workflow with id %q in namespace %q", signal.WorkflowID(objectID), signal.Namespace())
	}

	donechan := ctx.Done()
	workflow.Go(ctx, func(ctx workflow.Context) {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(workflow.GetSignalChannel(ctx, listener.SignalName), func(schan workflow.ReceiveChannel, closed bool) {
			val := new(eventloop.SignalDoneMessage)
			schan.Receive(ctx, val)
			if val != nil {
				set.Set(val.Result, val.Error)
			} else if closed {
				err := errors.New("notification signal channel was closed")
				set.Set(nil, err)
			} else {
				set.Set(nil, errors.New("should be unreachable - signal channel returned nothing without being closed"))
			}
		})

		selector.AddReceive(donechan, func(_ workflow.ReceiveChannel, _ bool) {
			set.SetError(ctx.Err())

			// Propagate the cancellation to the signalled workflow
			dctx, _ := workflow.NewDisconnectedContext(ctx)
			err := workflow.SignalExternalWorkflow(
				workflow.WithWorkflowNamespace(ctx, signal.Namespace()),
				signal.WorkflowID(objectID),
				"",
				CancelChannelName,
				signal,
			).Get(dctx, nil)
			if err != nil {
				e.mw.Incr("event_loop.signal", metrics.ToStatusTag("unable_to_send"))
				l.Error("unable to dispatch cancellation signal to workflow",
					zap.String("from-workflow", info.WorkflowExecution.ID),
					zap.String("to-workflow", objectID),
					zap.Error(err),
				)
			}
		})

		selector.Select(ctx)
	})

	return fut, nil
}

func (e *evClient) SendAndWait(ctx workflow.Context, id string, signal eventloop.Signal) error {
	var err error
	fut, err := e.SendAsync(ctx, id, signal)
	if err != nil {
		return err
	}

	return fut.Get(ctx, nil)
}
