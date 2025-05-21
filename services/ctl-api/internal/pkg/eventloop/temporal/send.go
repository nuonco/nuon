package temporal

import (
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
	var err error
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

// func (e *evClient) Cancel(wCtx workflow.Context, namespace, id string) {
// 	//
// }

func (e *evClient) SendAsync(ctx workflow.Context, id string, signal eventloop.Signal) (workflow.Future, error) {
	l := temporalzap.GetWorkflowLogger(ctx)
	if err := e.v.Struct(signal); err != nil {
		e.mw.Incr("event_loop.signal", metrics.ToStatusTag("invalid signal"))
		l.Error("invalid signal", zap.Error(err))
		return nil, errors.Wrap(err, "invalid signal")
	}

	if err := signal.PropagateContext(ctx); err != nil {
		e.mw.Incr("event_loop.signal", metrics.ToStatusTag("unable to propagate"))
		l.Error("unable to propagate", zap.Error(err))
		return nil, errors.Wrap(err, "unable to propagate values from context")
	}

	// TODO(sdboyer) should we handle start/restart cases here, like the other Send methods?

	info := workflow.GetInfo(ctx)
	listener := eventloop.SignalListener{
		WorkflowID: info.WorkflowExecution.ID,
		Namespace:  info.Namespace,
		SignalName: shortid.NewNanoID("listen"),
	}

	if err := eventloop.AppendListenerIDs(signal, listener); err != nil {
		e.mw.Incr("event_loop.signal", metrics.ToStatusTag("unable to add listeners"))
		l.Error("could not register signal listeners", zap.Error(err))
		return nil, errors.Wrap(err, "unable to add listeners")
	}

	err := workflow.SignalExternalWorkflow(
		workflow.WithWorkflowNamespace(ctx, signal.Namespace()),
		id,
		"",
		signal.Name(),
		signal,
	)
	if err != nil {
		e.mw.Incr("event_loop.signal", metrics.ToStatusTag("unable_to_send"))
	}

	fut, set := workflow.NewFuture(ctx)

	workflow.Go(ctx, func(ctx workflow.Context) {
		val := new(eventloop.SignalDoneMessage)
		schan := workflow.GetSignalChannel(ctx, listener.SignalName)
		closed := schan.Receive(ctx, val)

		if val != nil {
			set.Set(val.Result, val.Error)
		} else if closed {
			err := errors.New("notification signal channel was closed")
			set.Set(nil, err)
		} else {
			err := errors.New("should be unreachable - signal channel returned nothing without being closed")
			set.Set(nil, err)
		}
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
