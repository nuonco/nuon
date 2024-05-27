package worker

import (
	"errors"
	"strconv"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
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

		var err error
		switch signal.SignalType() {
		case signals.OperationCreated:
			op = "created"
			err = w.created(ctx, req.ID)
			if err != nil {
				status = "error"
				l.Error("unable to handle created signal", zap.Error(err))
				return
			}
		case signals.OperationPollDependencies:
			op = "poll_dependencies"
			err = w.pollDependencies(ctx, req.ID)
			if err != nil {
				status = "error"
				l.Error("unable to poll dependencies", zap.Error(err))
				return
			}
		case signals.OperationProvision:
			op = "provision"
			err = w.provision(ctx, req.ID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to provision", zap.Error(err))
				return
			}
		case signals.OperationReprovision:
			op = "reprovision"
			err = w.reprovision(ctx, req.ID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to reprovision", zap.Error(err))
				return
			}
		case signals.OperationDelete:
			op = "delete"
			err = w.delete(ctx, req.ID, req.SandboxMode)
			if err != nil {
				status = "delete"
				l.Error("unable to delete", zap.Error(err))
				return
			}
			finished = true
		case signals.OperationDeprovision:
			op = "deprovision"
			err = w.deprovision(ctx, req.ID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to deprovision", zap.Error(err))
				return
			}
		case signals.OperationForgotten:
			op = "forgotten"
			err = w.forget(ctx, req.ID)
			if err != nil {
				l.Error("unable to forget", zap.Error(err))
			}
			finished = true
		case signals.OperationDeployComponents:
			op = "deploy_components"
			err = w.deployComponents(ctx, req.ID, req.SandboxMode, signal.Async)
			if err != nil {
				status = "error"
				l.Error("unable to queue deploys for components", zap.Error(err))
				return
			}
		case signals.OperationTeardownComponents:
			op = "teardown_components"
			err = w.teardownComponents(ctx, req.ID, req.SandboxMode, signal.Async)
			if err != nil {
				status = "error"
				l.Error("unable to queue teardown deploys for components", zap.Error(err))
				return
			}
		case signals.OperationDeploy:
			op = "deploy"
			err = w.deploy(ctx, req.ID, signal.DeployID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to deploy", zap.Error(err))
				return
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
