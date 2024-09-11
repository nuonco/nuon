package worker

import (
	"errors"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest) error {
	defaultTags := map[string]string{"sandbox_mode": "false"}
	w.mw.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags, "op", "started")...)

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, signals.TemporalNamespace)
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, _ bool) {
		var signal signals.Signal
		channelOpen := channel.Receive(ctx, &signal)
		if !channelOpen {
			w.logger.Info("channel was closed")
			return
		} else {
			w.logger.Info("channel is open.")
		}

		if err := signal.Validate(w.v); err != nil {
			w.logger.Info("invalid signal", zap.Error(err))
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
			err := w.created(ctx)
			if err != nil {
				status = "error"
				w.logger.Error("unable to handle created signal", zap.Error(err))
				return
			}

		case signals.OperationReconcile:
			op = "reconcile"
			err := w.reconcile(ctx)
			if err != nil {
				status = "error"
				w.logger.Error("unable to handle reconcile signal", zap.Error(err))
				return
			}

		case signals.OperationRestart:
			op = "restart"
			err := w.restart(ctx)
			if err != nil {
				status = "error"
				w.logger.Error("unable to handle restart signal", zap.Error(err))
				return
			}

		}
	})

	for !finished {
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			w.mw.Incr(ctx, "event_loop.canceled", metrics.ToTags(defaultTags)...)
			w.logger.Error("workflow canceled")
			break
		}
		selector.Select(ctx)
	}

	w.mw.Incr(ctx, "event_loop.finish", metrics.ToTags(defaultTags)...)
	return nil
}
