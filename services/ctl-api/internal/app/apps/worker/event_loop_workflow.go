package worker

import (
	"errors"
	"strconv"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest) error {
	defaultTags := map[string]string{"sandbox_mode": strconv.FormatBool(req.SandboxMode)}
	w.mw.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags, "op", "started")...)
	l := workflow.GetLogger(ctx)

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

		switch signal.SignalType() {
		case signals.OperationCreated:
			op = "created"
			if err := w.created(ctx, req.ID); err != nil {
				status = "error"
				l.Info("unable to handle created signal: %w", zap.Error(err))
			}
		case signals.OperationPollDependencies:
			op = "poll_dependencies"
			if err := w.pollDependencies(ctx, req.ID); err != nil {
				status = "error"
				l.Info("unable to poll app dependencies: %w", zap.Error(err))
			}
		case signals.OperationProvision:
			op = "provision"
			if err := w.provision(ctx, req.ID, req.SandboxMode); err != nil {
				status = "error"
				l.Info("unable to provision app: %w", zap.Error(err))
			}
		case signals.OperationReprovision:
			op = "reprovision"
			if err := w.reprovision(ctx, req.ID, req.SandboxMode); err != nil {
				status = "error"
				l.Info("unable to reprovision app: %w", zap.Error(err))
			}
		case signals.OperationUpdateSandbox:
			op = "update_sandbox"
			if err := w.updateSandbox(ctx, req.ID, "", req.SandboxMode); err != nil {
				status = "update_sandbox"
				l.Info("unable to provision app: %w", zap.Error(err))
			}
		case signals.OperationDeprovision:
			op = "deprovision"
			if err := w.deprovision(ctx, req.ID, req.SandboxMode); err != nil {
				status = "deprovision"
				l.Info("unable to deprovision app: %w", zap.Error(err))
				return
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
