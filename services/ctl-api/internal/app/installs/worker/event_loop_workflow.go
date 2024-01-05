package worker

import (
	"errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/signals"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) InstallEventLoop(ctx workflow.Context, req signals.InstallEventLoopRequest) error {
	l := workflow.GetLogger(ctx)

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, req.InstallID)
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

		var err error
		switch signal.Operation {
		case signals.OperationPollDependencies:
			err = w.pollDependencies(ctx, req.InstallID)
			if err != nil {
				l.Error("unable to poll dependencies", zap.Error(err))
				return
			}
		case signals.OperationProvision:
			err = w.provision(ctx, req.InstallID, req.SandboxMode)
			if err != nil {
				l.Error("unable to provision", zap.Error(err))
				return
			}
		case signals.OperationReprovision:
			err = w.reprovision(ctx, req.InstallID, req.SandboxMode)
			if err != nil {
				l.Error("unable to reprovision", zap.Error(err))
				return
			}
		case signals.OperationDelete:
			err = w.delete(ctx, req.InstallID, req.SandboxMode)
			if err != nil {
				l.Error("unable to delete", zap.Error(err))
				return
			}
			finished = true
		case signals.OperationDeprovision:
			err = w.deprovision(ctx, req.InstallID, req.SandboxMode)
			if err != nil {
				l.Error("unable to deprovision", zap.Error(err))
				return
			}
		case signals.OperationForgotten:
			finished = true
		case signals.OperationDeployComponents:
			err = w.deployComponents(ctx, req.InstallID, req.SandboxMode, signal.Async)
			if err != nil {
				l.Error("unable to queue deploys for components", zap.Error(err))
				return
			}
		case signals.OperationTeardownComponents:
			err = w.teardownComponents(ctx, req.InstallID, req.SandboxMode, signal.Async)
			if err != nil {
				l.Error("unable to queue teardown deploys for components", zap.Error(err))
				return
			}
		case signals.OperationDeploy:
			err = w.deploy(ctx, req.InstallID, signal.DeployID, req.SandboxMode)
			if err != nil {
				l.Error("unable to deploy", zap.Error(err))
				return
			}
		}
	})

	for !finished {
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			l.Error("workflow canceled")
			break
		}

		selector.Select(ctx)
	}

	return nil
}
