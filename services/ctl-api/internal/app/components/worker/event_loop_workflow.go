package worker

import (
	"errors"
	"strconv"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/signals"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) ComponentEventLoop(ctx workflow.Context, req signals.ComponentEventLoopRequest) error {
	defaultTags := map[string]string{"sandbox_mode": strconv.FormatBool(req.SandboxMode)}
	w.mw.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags, "op", "started")...)
	l := zap.L()

	finished := false

	signalChan := workflow.GetSignalChannel(ctx, req.ComponentID)
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

		startTS := workflow.Now(ctx)
		op := ""
		status := "ok"
		defer func() {
			tags := generics.MergeMap(map[string]string{
				"op":     op,
				"status": status,
			}, defaultTags)
			dur := workflow.Now(ctx).Sub(startTS)

			w.mw.Timing(ctx, "event_loop.signal_duration", dur, metrics.ToTags(tags)...)
			w.mw.Incr(ctx, "event_loop.signal", metrics.ToTags(tags)...)
		}()

		switch signal.Operation {
		case signals.OperationPollDependencies:
			op = "poll_dependencies"
			if err := w.pollDependencies(ctx, req.ComponentID); err != nil {
				status = "error"
				l.Info("unable to poll app status for readiness: %w", zap.Error(err))
			}
		case signals.OperationProvision:
			op = "provision"
			if err := w.provision(ctx, req.ComponentID); err != nil {
				status = "error"
				l.Info("unable to provision: %w", zap.Error(err))
			}
		case signals.OperationQueueBuild:
			op = "queue_build"
			if err := w.queueBuild(ctx, req.ComponentID); err != nil {
				status = "error"
				l.Info("unable to handle queue build: %w", zap.Error(err))
			}
		case signals.OperationBuild:
			op = "build"
			if err := w.build(ctx, req.ComponentID, signal.BuildID, req.SandboxMode); err != nil {
				status = "error"
				l.Info("unable to build component: %w", zap.Error(err))
			}
		case signals.OperationDelete:
			op = "delete"
			if err := w.delete(ctx, req.ComponentID, req.SandboxMode); err != nil {
				status = "error"
				l.Info("unable to delete component: %w", zap.Error(err))
			}
			finished = true
		}
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
