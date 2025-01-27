package worker

import (
	"errors"
	"fmt"
	"strconv"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (w *Workflows) handleSignal(ctx workflow.Context, wkflowReq eventloop.EventLoopRequest, signal signals.Signal, defaultTags map[string]string) (bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	ctx = signal.GetWorkflowContext(ctx)
	finished := false
	startTS := workflow.Now(ctx)

	// NOTE(jm): this wrapper type is basically the result of two requirements:
	// 1. - to pass state for the org/sandbox-mode into the signal handler
	// 2. - generate a workflow id for the signal handler
	//
	// However, neither of these should be requirements because a.) it's preferable to be more explicit vs magically
	// generating the signal workflow id, and b.) we use context propagation + state on the workflow type to pass
	// the other information in.
	//
	// TODO(jm): temporal-gen does not currently support this, but I think the workflow-id should be an argument on
	// the generated function, ie: these methods should be w.AwaitCreated(wkflow-id, signal) vs
	// w.AwaitCreated(sigReq)

	sigReq := signals.RequestSignal{
		Signal:           &signal,
		EventLoopRequest: wkflowReq,
	}

	switch signal.SignalType() {

	// basic signals for handling crud operations and bootup
	case signals.OperationCreated:
		err = w.AwaitCreated(ctx, sigReq)
	case signals.OperationRestart:
		err = w.AwaitRestart(ctx, sigReq)
	case signals.OperationPollDependencies:
		err = w.AwaitPollDependencies(ctx, sigReq)
	case signals.OperationDelete:
		err = w.AwaitDelete(ctx, sigReq)

		// actual execution signals
	case signals.OperationConfigCreated:
		err = w.AwaitConfigCreated(ctx, sigReq)
	default:
		err = fmt.Errorf("unhandled signal %s", signal.SignalType())
	}

	op := signal.Name()
	status := "ok"
	if err != nil {
		l.Error("signal handling failed", zap.Error(err), zap.String("signal", op))
		status = "error"
	} else {
		l.Debug("signal handling succeeded", zap.String("signal", op))
	}

	// write out metrics
	tags := generics.MergeMap(map[string]string{
		"op":     op,
		"status": status,
	}, defaultTags)

	dur := workflow.Now(ctx).Sub(startTS)

	w.mw.Timing(ctx, "event_loop.signal_duration", dur, metrics.ToTags(tags)...)
	w.mw.Incr(ctx, "event_loop.signal", metrics.ToTags(tags)...)

	return finished, nil
}

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	defaultTags := map[string]string{"sandbox_mode": strconv.FormatBool(req.SandboxMode)}
	w.mw.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags, "op", "started")...)

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, req.ID)
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, _ bool) {
		var signal signals.Signal
		channelOpen := channel.Receive(ctx, &signal)
		if !channelOpen {
			l.Info("channel was closed")
			return
		}

		if err := signal.Validate(w.v); err != nil {
			l.Info("invalid signal", zap.Error(err))
		}

		finished, err = w.handleSignal(ctx, req, signal, defaultTags)
	})

	for !finished {
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			w.mw.Incr(ctx, "event_loop.canceled", metrics.ToTags(defaultTags)...)
			l.Error("workflow canceled")
			break
		}

		selector.Select(ctx)
	}

	w.mw.Incr(ctx, "event_loop.finish", metrics.ToTags(defaultTags)...)
	return nil
}
