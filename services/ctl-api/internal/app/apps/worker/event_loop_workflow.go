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
	EventLoopWorkflowName string = "AppEventLoop"
)

func EventLoopWorkflowID(appID string) string {
	return fmt.Sprintf("%s-event-loop", appID)
}

type AppEventLoopRequest struct {
	AppID       string
	SandboxMode bool
}

func (w *Workflows) AppEventLoop(ctx workflow.Context, req AppEventLoopRequest) error {
	defaultTags := map[string]string{"sandbox_mode": strconv.FormatBool(req.SandboxMode)}
	w.mw.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags, "op", "started")...)
	l := workflow.GetLogger(ctx)

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, req.AppID)
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
			w.mw.Incr(ctx, "event_loop.signal", metrics.ToTags(tags)...)
		}()

		switch signal.Operation {
		case OperationPollDependencies:
			op = "poll_dependencies"
			if err := w.pollDependencies(ctx, req.AppID); err != nil {
				status = "error"
				l.Info("unable to poll app dependencies: %w", zap.Error(err))
			}
		case OperationProvision:
			op = "provision"
			if err := w.provision(ctx, req.AppID, req.SandboxMode); err != nil {
				status = "error"
				l.Info("unable to provision app: %w", zap.Error(err))
			}
		case OperationReprovision:
			op = "reprovision"
			if err := w.reprovision(ctx, req.AppID, req.SandboxMode); err != nil {
				status = "error"
				l.Info("unable to reprovision app: %w", zap.Error(err))
			}
		case OperationUpdateSandbox:
			op = "update_sandbox"
			if err := w.updateSandbox(ctx, req.AppID, signal.SandboxReleaseID, req.SandboxMode); err != nil {
				status = "update_sandbox"
				l.Info("unable to provision app: %w", zap.Error(err))
			}
		case OperationConfigCreated:
			if err := w.syncConfig(ctx, req.AppID, signal.AppConfigID, req.SandboxMode); err != nil {
				l.Info("unable to sync config: %w", zap.Error(err))
			}
		case OperationDeprovision:
			op = "deprovision"
			if err := w.deprovision(ctx, req.AppID, req.SandboxMode); err != nil {
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
