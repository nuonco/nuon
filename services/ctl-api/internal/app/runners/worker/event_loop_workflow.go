package worker

import (
	"errors"
	"strconv"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest) error {
	defaultTags := map[string]string{"sandbox_mode": strconv.FormatBool(req.SandboxMode)}
	w.mw.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags, "op", "started")...)
	l := zap.L()

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

		sreq := signals.RequestSignal{
			Signal:           &signal,
			EventLoopRequest: req,
		}

		switch signal.SignalType() {

		// signals to handle operations of the runner
		case signals.OperationRestart:
			op = "restart"
			if err := w.AwaitRestart(ctx, sreq); err != nil {
				status = "error"
				l.Info("unable to handle restart signal: %w", zap.Error(err))
			}
		case signals.OperationDelete:
			op = "delete"
			if err := w.AwaitDelete(ctx, sreq); err != nil {
				status = "error"
				l.Info("unable to handle delete signal: %w", zap.Error(err))
			}
		case signals.OperationForceDelete:
			op = "force_delete"
			if err := w.AwaitForce_delete(ctx, sreq); err != nil {
				status = "error"
				l.Info("unable to handle force delete signal: %w", zap.Error(err))
			}
		case signals.OperationCreated:
			op = "created"
			if err := w.AwaitCreated(ctx, sreq); err != nil {
				status = "error"
				l.Info("unable to handle created signal: %w", zap.Error(err))
			}
		case signals.OperationProvision:
			op = "provision"
			if err := w.AwaitProvision(ctx, sreq); err != nil {
				status = "error"
				l.Info("unable to handle provision signal: %w", zap.Error(err))
			}
		case signals.OperationReprovision:
			op = "reprovision"
			if err := w.AwaitReprovision(ctx, sreq); err != nil {
				status = "error"
				l.Info("unable to handle reprovision signal: %w", zap.Error(err))
			}
		case signals.OperationDeprovision:
			op = "deprovision"
			if err := w.AwaitDeprovision(ctx, sreq); err != nil {
				status = "error"
				l.Info("unable to handle deprovision signal: %w", zap.Error(err))
			}

			// operations to handle the basics
			// NOTE(jm): not all jobs have to be queued, if they are created
			// with the "available" status,
			// such as a health check, they will immediately be picked up
		case signals.OperationJobQueued:
			op = "job_queued"
			if err := w.processJob(ctx, req.ID, signal.JobID); err != nil {
				status = "error"
				l.Info("unable to handle queued build: %w", zap.Error(err))
			}
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
