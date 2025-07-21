package loop

import (
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (w *Loop[SignalType, ReqSig]) handleSignal(pctx workflow.Context, wkflowReq eventloop.EventLoopRequest, signal SignalType, defaultTags map[string]string) error {
	l, err := log.WorkflowLogger(pctx)
	if err != nil {
		return err
	}

	ctx := signal.GetWorkflowContext(w.cb.Join(pctx, func(other SignalType) bool {
		return w.signalsMatch(signal, other)
	}))

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

	startTS := workflow.Now(ctx)
	sigReq := w.NewRequestSignal(wkflowReq, signal)
	handler, ok := w.Handlers[signal.SignalType()]
	if !ok || handler == nil {
		err = fmt.Errorf("unhandled signal %s", signal.SignalType())
	} else {
		err = handler(ctx, sigReq)
	}

	op := signal.Name()
	status := "ok"
	if err != nil {
		status = "error"
		if temporal.IsTimeoutError(err) {
			status = "timeout"
		} else if temporal.IsPanicError(err) {
			status = "panic"
		}
		l.Error("signal handling failed", zap.Error(err), zap.String("signal", op))
	} else {
		l.Debug("signal handling succeeded", zap.String("signal", op))
	}

	// write out metrics
	tags := generics.MergeMap(map[string]string{
		"op":     op,
		"status": status,
	}, defaultTags)

	dur := workflow.Now(ctx).Sub(startTS)

	w.MW.Timing(ctx, "event_loop.signal_duration", dur, metrics.ToTags(tags)...)
	w.MW.Incr(ctx, "event_loop.signal", metrics.ToTags(tags)...)

	for _, listener := range signal.Listeners() {
		l.Debug("notifying signal listener",
			zap.String("listener-workflow-id", listener.WorkflowID),
			zap.String("listener-namespace", listener.Namespace),
			zap.String("signal", op),
		)

		lctx := workflow.WithWorkflowNamespace(ctx, listener.Namespace)
		lerr := workflow.SignalExternalWorkflow(
			lctx,
			listener.WorkflowID,
			"",
			listener.SignalName,
			eventloop.SignalDoneMessage{
				// CompletedSignal: signal,
				Error:  err,
				Result: nil,
			},
		).Get(lctx, nil)
		if lerr != nil {
			// Calling workflow may have gone away, so don't worry about these too much.
			l.Warn("error notifying signal listener",
				zap.Error(lerr),
				zap.String("listener-workflow-id", listener.WorkflowID),
				zap.String("listener-namespace", listener.Namespace),
				zap.String("signal", op),
			)

			w.MW.Incr(ctx, "event_loop.signal.notify_errors", metrics.ToTags(tags)...)
		}
	}

	ldur := workflow.Now(ctx).Sub(startTS) - dur
	w.MW.Timing(ctx, "event_loop.signal.notify_duration", ldur, metrics.ToTags(tags)...)

	return nil
}

func (w *Loop[SignalType, ReqSig]) signalsMatch(a, b SignalType) bool {
	if a.Name() != b.Name() {
		return false
	}

	// Allow cancellation only if the signal name of the listener matches
	//
	// TODO(sdboyer) this is a bit of a hack, but it's an effective way of ensuring for now that ONLY the workflow that called
	// SendAsync can trigger cancellation this way, because SendAsync generates these listener ids on the fly.
	// This may be more restrictive than the contract we eventually want, but right now, it has no false positives.
	for _, alistener := range a.Listeners() {
		for _, blistener := range b.Listeners() {
			if alistener.SignalName == blistener.SignalName {
				return true
			}
		}
	}
	return false
}
