package loop

import (
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	sharedactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/activities"
)

const (
	maxSignals int = 10
)

type Loop[T eventloop.Signal, R any] struct {
	Cfg              *internal.Config
	MW               tmetrics.Writer
	V                *validator.Validate
	Handlers         map[eventloop.SignalType]func(workflow.Context, R) error
	NewRequestSignal func(eventloop.EventLoopRequest, T) R
}

func (w *Loop[T, R]) handleSignal(ctx workflow.Context, wkflowReq eventloop.EventLoopRequest, signal T, defaultTags map[string]string) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	ctx = signal.GetWorkflowContext(ctx)
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

	w.MW.Timing(ctx, "event_loop.signal_duration", dur, metrics.ToTags(tags)...)
	w.MW.Incr(ctx, "event_loop.signal", metrics.ToTags(tags)...)

	return nil
}

func (w *Loop[T, R]) Run(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []T) error {
	version, err := sharedactivities.AwaitGetVersion(ctx, sharedactivities.GetVersionRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to get version")
	}

	ctx = cctx.SetVersionWorkflowContext(ctx, version)

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	defaultTags := map[string]string{"sandbox_mode": strconv.FormatBool(req.SandboxMode)}
	w.MW.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags, "op", "started")...)

	signalChan := workflow.GetSignalChannel(ctx, req.ID)

	// handle any pending signals
	for _, pendingSignal := range pendingSignals {
		err := w.handleSignal(ctx, req, pendingSignal, defaultTags)
		if err != nil {
			l.Error("error handling signal", zap.Error(err))
			continue
		}

		if pendingSignal.Stop() {
			w.MW.Incr(ctx, "event_loop.finish", metrics.ToTags(defaultTags)...)
			return nil
		}

		// NOTE: if a restart is in here, we do not actually restart the loop, since this is the initialization,
		// anyway.
	}

	// execute the event loop
	pendingSignals = make([]T, 0)
	signalCount := 0
	stop := false
	restart := false
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, _ bool) {
		var signal T
		channelOpen := channel.Receive(ctx, &signal)
		if !channelOpen {
			l.Info("channel was closed")
			return
		}

		// restart requires the signal to be handed on the fresh loop
		if signal.Restart() {
			restart = true
			pendingSignals = append(pendingSignals, signal)
			return
		}

		// version is not valid, so we can not attempt the job
		if w.Cfg.Version != version {
			restart = true
			pendingSignals = append(pendingSignals, signal)
			return
		}

		if err := signal.Validate(w.V); err != nil {
			l.Info("invalid signal", zap.Error(err))
		}

		signalCount += 1
		err = w.handleSignal(ctx, req, signal, defaultTags)
		if err != nil {
			l.Error("error handling signal", zap.Error(err))
			return
		}

		stop = signal.Stop()
	})

	for !stop {
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			w.MW.Incr(ctx, "event_loop.canceled", metrics.ToTags(defaultTags)...)
			l.Error("workflow canceled")
			break
		}

		selector.Select(ctx)

		if restart {
			w.MW.Incr(ctx, "event_loop.restarted", metrics.ToTags(defaultTags)...)
			l.Error("workflow restarted")
			req.RestartCount += 1
			pendingSignals = append(pendingSignals, w.drainSignals(signalChan)...)
			return workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, req, pendingSignals)
		}

		if signalCount >= maxSignals {
			w.MW.Incr(ctx, "event_loop.max_signals", metrics.ToTags(defaultTags)...)
			l.Error("workflow hit max signals")
			req.RestartCount += 1
			pendingSignals = append(pendingSignals, w.drainSignals(signalChan)...)
			return workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, req, pendingSignals)
		}
	}

	w.MW.Incr(ctx, "event_loop.finish", metrics.ToTags(defaultTags)...)
	return nil
}
