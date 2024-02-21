package worker

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	EventLoopWorkflowName string = "ReleaseEventLoop"
)

func EventLoopWorkflowID(releaseID string) string {
	return fmt.Sprintf("%s-event-loop", releaseID)
}

type ReleaseEventLoopRequest struct {
	ReleaseID   string
	SandboxMode bool
}

func (w *Workflows) ReleaseEventLoop(ctx workflow.Context, req ReleaseEventLoopRequest) error {
	defaultTags := map[string]string{"sandbox_mode": strconv.FormatBool(req.SandboxMode)}
	w.mw.Incr(ctx, "event_loop.start", 1, metrics.ToTags(defaultTags, "op", "started")...)
	l := workflow.GetLogger(ctx)

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, req.ReleaseID)
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, _ bool) {
		var signal Signal
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
			w.mw.Incr(ctx, "event_loop.signal", 1, metrics.ToTags(tags)...)
		}()

		switch signal.Operation {
		case OperationPollDependencies:
			op = "poll_dependencies"
			if err := w.pollDependencies(ctx, req.ReleaseID); err != nil {
				status = "error"
				l.Info("unable to poll dependencies: %w", zap.Error(err))
			}
		case OperationProvision:
			op = "provision"
			if err := w.provision(ctx, req.ReleaseID, req.SandboxMode); err != nil {
				status = "error"
				l.Info("unable to provision release: %w", zap.Error(err))
			}
			finished = true
		}
	})
	for !finished {
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			w.mw.Incr(ctx, "event_loop.canceled", 1, metrics.ToTags(defaultTags)...)
			l.Error("workflow canceled")
			break
		}

		selector.Select(ctx)
	}

	w.mw.Incr(ctx, "event_loop.finish", 1, metrics.ToTags(defaultTags)...)
	return nil
}
