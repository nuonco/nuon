package loop

import (
	"strconv"

	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/errs"
	"github.com/powertoolsdev/mono/pkg/metrics"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	ntemporal "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
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

	// the cancelbroker manages cancellations for child workflows, including holding signals for not-yet-received
	// signals
	cb *ntemporal.CancelBroker[SignalType]

	// conc holds the active concurrency groups within which signals are queued.
	conc map[string]*concgroup[SignalType]

	// concwg is a WaitGroup that tracks the completion of all spawned coroutines
	concwg workflow.WaitGroup

	// Each channel is used to broadcast that a distinct stopping conditions for the loop has been reached
	stopChan, restartChan, countChan workflow.Channel
	// Bools representing stopping conditions for the loop
	stop, restart, countExceeded, notexist bool

	// existsErr is set if the ExistsHook returns an error
	existsErr error

	// pendingSignals holds signals to be passed back in to a ContinueAsNew instance of this event loop
	pendingSignals []SignalType

	// The number of signals that have been ingested by the event loop. When exceeded, the loop will restart and ContinueAsNew
	sigcount int

	// hooks

	// StartupHook is called once, before the event loop starts in a call to Run
	StartupHook func(workflow.Context, eventloop.EventLoopRequest) error

	// ExistsHook is called before each signal is processed to check if the
	// underlying object exists. The check is skipped if no implementation is
	// provided.
	ExistsHook func(workflow.Context, eventloop.EventLoopRequest) (bool, error)
}

func (l *Loop[SignalType, ReqSig]) RunWithConcurrency(ctx workflow.Context, req eventloop.EventLoopRequest, incomingPendingSignals []SignalType) error {
	defaultTags := l.init(ctx, req)
	selector := l.createSelector(ctx, req, defaultTags)
	log, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	if l.StartupHook != nil {
		if err := l.StartupHook(ctx, req); err != nil {
			log.Error("startup hook failed, restarting", zap.Error(err))
			l.MW.Incr(ctx, "event_loop.restart", metrics.ToTags(defaultTags, "op", "restarted", "reason", "startup_hook_failure")...)
			req.RestartCount += 1
			return workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, req, incomingPendingSignals)
		}
	}

	// handle any pending signals
	l.handlePending(ctx, req, defaultTags, incomingPendingSignals)

	// Main loop. The select has several branches to unblock, primarily when a signal is received, but also
	// if any of our four stopping conditions are met:
	//
	//  - A signal with Stop() = true is received. No new signals will be ingested, but ones already in their concgroup queues will be processed before stopping.
	//  - The underlying object goes away. Everything shuts down right away, and both waiting and ingested signals are dropped.
	//  - A signal is sent with Restart(). No new signals will be ingested. All signals currently in their concgroup queues will be assembled in order
	//    (within their conc group) into pendingSignals. The workflow will restart with ContinueAsNew, and pending signals processed in their respective conc groups.
	// 	  The Restart() signal itself is processed as part of pendingSignals on the refreshed loop.
	//  - The maxSignals limit is exceeded. Behavior is identical to the Restart() case, but slightly different telemetry is emitted.
	for !(l.stop || l.notexist) {
		// fmt.Println(l.stop, l.notexist, l.restart, l.countExceeded)
		selector.Select(ctx)

		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			l.MW.Incr(ctx, "event_loop.canceled", metrics.ToTags(defaultTags)...)
			log.Info("workflow canceled")
			break
		}

		if l.restart || l.countExceeded {
			if l.restart {
				l.MW.Incr(ctx, "event_loop.restarted", metrics.ToTags(defaultTags)...)
				log.Info("workflow restarted")
				// TODO(jm): this should be removed once we ensure all event loops have an active
				// `WorkerVersion` in their request chain
				req.Version = l.Cfg.Version
			} else {
				l.MW.Incr(ctx, "event_loop.max_signals", metrics.ToTags(defaultTags)...)
				log.Info("workflow hit max signals")
			}

			l.restartChan.Close()
			l.stopChan.Close()
			l.countChan.Close()
			l.concwg.Wait(ctx)

			req.RestartCount += 1
			// This will drain all pending signals from the main signal queue
			for selector.HasPending() {
				selector.Select(ctx)
			}
			// It's critical that we return immediately after draining, without doing anything that would cause
			// this coroutine to yield. Otherwise, additional signals may be received, which would then be dropped
			// in the ContinueAsNew transition.
			return workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, req, l.pendingSignals)
		}
	}

	l.MW.Incr(ctx, "event_loop.finish", metrics.ToTags(defaultTags)...)
	return nil
}

func (l *Loop[SignalType, ReqSig]) init(ctx workflow.Context, req eventloop.EventLoopRequest) map[string]string {
	l.conc = make(map[string]*concgroup[SignalType])
	l.concwg = workflow.NewWaitGroup(ctx)
	l.cb = ntemporal.NewCancelBroker[SignalType](ctx)
	l.pendingSignals = make([]SignalType, 0)

	wkflowInfo := workflow.GetInfo(ctx)
	defaultTags := map[string]string{
		"sandbox_mode": strconv.FormatBool(req.SandboxMode),
		"namespace":    wkflowInfo.Namespace,
		"task_queue":   wkflowInfo.TaskQueueName,
	}
	l.MW.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags)...)
	l.MW.Gauge(ctx, "event_loop.restart_count", float64(req.RestartCount), metrics.ToTags(defaultTags)...)
	l.MW.Gauge(ctx, "event_loop.version_change_count", float64(req.VersionChangeCount), metrics.ToTags(defaultTags)...)

	return defaultTags
}

func (l *Loop[SignalType, ReqSig]) logger(ctx workflow.Context) *zap.Logger {
	log, err := log.WorkflowLogger(ctx)
	if err != nil {
		panic("unreachable - unable to get workflow logger after it was retrieved successfully in parent fn")
	}
	return log
}

// TODO(sdboyer) delete everything below here once we're ready to fully adopt concurrent event loops

func (w *Loop[SignalType, ReqSig]) Run(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []SignalType) error {
	signalChan := workflow.GetSignalChannel(ctx, req.ID)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}
	w.cb = ntemporal.NewCancelBroker[SignalType](ctx)

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

	if w.StartupHook != nil && !isWorkflowELoop(wkflowInfo.WorkflowExecution.ID) {
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

func (l *Loop[SignalType, ReqSig]) drainSignals(ch workflow.ReceiveChannel) []SignalType {
	var signals []SignalType
	for {
		var signal SignalType
		ok := ch.ReceiveAsync(&signal)
		if !ok {
			break
		}

		signals = append(signals, signal)
	}

	return signals
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

// TODO(sdboyer) remove this - a temporary hack while user workflows are still cohabiting with event loops
func isWorkflowELoop(id string) bool {
	return strings.Index(id, "event-loop-workflow-") == 0
}
