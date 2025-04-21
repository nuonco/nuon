package loop

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/errs"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

const (
	maxSignals          int = 10
	checkExistsInterval     = 24 * time.Hour * 7
)

// Loop is the generic implementation of Nuon's Temporal-backed event loop. Start an event
// loop by creating an instance of Loop, then calling [*Loop.Run].
//
// Each event loop is bound to a single domain of signals. The signal domain is defined by
// the SignalType constraint.
//
// ReqSig is the type of the request signal used while processing signals. It is derived from
// from the SignalType.
// TODO(sdboyer): refactor to get rid of the need for the explicit request signal type
type Loop[SignalType eventloop.Signal, ReqSig any] struct {
	Cfg              *internal.Config
	MW               tmetrics.Writer
	V                *validator.Validate
	Handlers         map[eventloop.SignalType]func(workflow.Context, ReqSig) error
	NewRequestSignal func(eventloop.EventLoopRequest, SignalType) ReqSig

	// hooks

	// StartupHook is called once, before the event loop starts in a call to Run
	StartupHook func(workflow.Context, eventloop.EventLoopRequest) error

	// checkExistsHook is called before each signal is processed to check if the
	// underlying object exists. The check is skipped if no implementation is
	// provided.
	ExistsHook func(workflow.Context, eventloop.EventLoopRequest) (bool, error)
}

func (w *Loop[SignalType, ReqSig]) handleSignal(ctx workflow.Context, wkflowReq eventloop.EventLoopRequest, signal SignalType, defaultTags map[string]string) error {
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
		status = "error"
		if temporal.IsPanicError(err) {
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

	switch status {
	case "error", "panic":
		errs.ReportToSentry(err, &errs.SentryErrOptions{
			Tags: tags,
		})
	}

	dur := workflow.Now(ctx).Sub(startTS)

	w.MW.Timing(ctx, "event_loop.signal_duration", dur, metrics.ToTags(tags)...)
	w.MW.Incr(ctx, "event_loop.signal", metrics.ToTags(tags)...)

	var listenErrs []error
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
		).Get(lctx, nil) // TODO(sdboyer) double check that this future just waits for the signal to be created, not processed
		if lerr != nil {
			l.Error("error notifying signal listener",
				zap.Error(lerr),
				zap.String("listener-workflow-id", listener.WorkflowID),
				zap.String("listener-namespace", listener.Namespace),
				zap.String("signal", op),
			)
			listenErrs = append(listenErrs, lerr)

			w.MW.Incr(ctx, "event_loop.signal.notify_errors", metrics.ToTags(tags)...)
		}
	}

	ldur := workflow.Now(ctx).Sub(startTS) - dur
	w.MW.Timing(ctx, "event_loop.signal.notify_duration", ldur, metrics.ToTags(tags)...)

	if len(listenErrs) > 0 {
		return errors.Wrap(errors.Join(listenErrs...), "error notifying signal listeners")
	}

	return nil
}

func (w *Loop[SignalType, ReqSig]) Run(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []SignalType) error {
	signalChan := workflow.GetSignalChannel(ctx, req.ID)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	wkflowInfo := workflow.GetInfo(ctx)
	pendingSignals = w.filterSignals(ctx, wkflowInfo, pendingSignals)

	defaultTags := map[string]string{
		"sandbox_mode": strconv.FormatBool(req.SandboxMode),
		"namespace":    wkflowInfo.Namespace,
		"task_queue":   wkflowInfo.TaskQueueName,
	}
	w.MW.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags)...)
	w.MW.Gauge(ctx, "event_loop.restart_count", float64(req.RestartCount), metrics.ToTags(defaultTags)...)
	w.MW.Gauge(ctx, "event_loop.version_change_count", float64(req.VersionChangeCount), metrics.ToTags(defaultTags)...)

	if w.StartupHook != nil {
		if err := w.StartupHook(ctx, req); err != nil {
			l.Error("startup hook failed, restarting", zap.Error(err))
			w.MW.Incr(ctx, "event_loop.restart", metrics.ToTags(defaultTags, "op", "restarted", "reason", "startup_hook_failure")...)
			req.RestartCount += 1
			return workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, req, pendingSignals)
		}
	}

	// handle any pending signals
	w.MW.Gauge(ctx, "event_loop.pending_signals", float64(len(pendingSignals)), metrics.ToTags(defaultTags)...)
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
	pendingSignals = make([]SignalType, 0)
	signalCount := 0
	stop := false
	restart := false
	checkExists := func() {
		exists := true
		var err error
		if w.ExistsHook != nil {
			exists, err = w.ExistsHook(ctx, req)
			if err != nil {
				// This should only be reachable in the event of an underlying temporal error.
				l.Error("error checking for existence of underlying object", zap.Error(err))
			}
		}
		stop = !exists
	}

	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, _ bool) {
		var signal SignalType
		channelOpen := channel.Receive(ctx, &signal)
		if !channelOpen {
			l.Info("channel was closed")
			return
		}

		checkExists()
		if stop {
			// By aborting here if the underlying object does not exist, we encode that no event loop-based cleanup may occur
			// after the underlying object is deleted.
			l.Warn("underlying object does not exist, shutting down event loop", zap.Error(err))
		} else {
			// restart requires the signal to be handed on the fresh loop
			if signal.Restart() {
				pendingSignals = append(pendingSignals, signal)
				restart = true
				return
			}

			if err := signal.Validate(w.V); err != nil {
				l.Error("invalid signal", zap.Error(err))
			}

			signalCount += 1
			err = w.handleSignal(ctx, req, signal, defaultTags)
			if err != nil {
				l.Error("error handling signal", zap.Error(err))
				return
			}

			stop = signal.Stop()
		}
	})

	existsFut := func(f workflow.Future) {
		checkExists()
	}

	for !stop {
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			w.MW.Incr(ctx, "event_loop.canceled", metrics.ToTags(defaultTags)...)
			l.Error("workflow canceled")
			break
		}

		// Register a new future on the selector with a fresh timeout that will fire
		// the exists check
		fsel := selector.AddFuture(workflow.NewTimer(ctx, checkExistsInterval), existsFut)
		fsel.Select(ctx)

		if restart {
			w.MW.Incr(ctx, "event_loop.restarted", metrics.ToTags(defaultTags)...)
			l.Error("workflow restarted")
			// TODO(jm): this should be removed once we ensure all event loops have an active
			// `WorkerVersion` in their request chain
			req.Version = w.Cfg.Version
			req.RestartCount += 1
			pendingSignals = append(pendingSignals, w.filterSignals(ctx, wkflowInfo, w.drainSignals(signalChan))...)
			return workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, req, pendingSignals)
		}

		if signalCount >= maxSignals {
			w.MW.Incr(ctx, "event_loop.max_signals", metrics.ToTags(defaultTags)...)
			l.Error("workflow hit max signals")
			req.RestartCount += 1
			pendingSignals = append(pendingSignals, w.filterSignals(ctx, wkflowInfo, w.drainSignals(signalChan))...)
			return workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, req, pendingSignals)
		}
	}

	w.MW.Incr(ctx, "event_loop.finish", metrics.ToTags(defaultTags)...)
	return nil
}

// This is a temporary (let's hope) function for identifying unpopulated/nil signals and reporting on them
func (w *Loop[SignalType, ReqSig]) validateSignal(ctx workflow.Context, info *workflow.Info, sig SignalType, depth int) bool {
	// send these in a defer, because nil signals will panic on these calls
	tags := map[string]string{
		"namespace":       info.Namespace,
		"task_queue":      info.TaskQueueName,
		"original_run_id": info.OriginalRunID,
	}

	var ret bool
	err := sig.Validate(w.V)

	// Getting info from the signal may panic if we have a nil, so set up a defer first
	defer func() {
		rec := recover()
		if rec != nil || err != nil {
			w.MW.Incr(ctx, "event_loop.invalid_signals", metrics.ToTags(tags)...)

			if rec != nil {
				// We might have both a validation error and a panic. Overwrite if we have the panic.
				if rerr, ok := rec.(error); ok {
					err = errors.WithStackDepth(errors.Wrap(rerr, "panic getting data from signal, likely nil"), depth)
				} else {
					err = errors.NewWithDepthf(depth, "panic, but recovery type was not an error: %s", rec)
				}
			} else {
				// panic handler supersedes the regular validation err
				err = errors.WithStackDepth(err, depth)
			}

			// TODO(sdboyer) replace this with something that goes to dd
			errs.ReportToSentry(err, &errs.SentryErrOptions{
				Tags: tags,
			})
		} else {
			ret = true
		}
	}()
	tags["workflow_type"] = sig.Name()

	ret = err == nil
	return ret // overridden by defer, no matter what
}

func (w *Loop[SignalType, ReqSig]) filterSignals(ctx workflow.Context, info *workflow.Info, sigs []SignalType) []SignalType {
	ret := make([]SignalType, 0, len(sigs))
	for _, sig := range sigs {
		if w.validateSignal(ctx, info, sig, 2) {
			ret = append(ret, sig)
		}
	}
	return ret
}
