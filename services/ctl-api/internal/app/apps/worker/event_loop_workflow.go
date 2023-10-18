package worker

import (
	"fmt"

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
	l := zap.L()

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

		switch signal.Operation {
		case OperationPollDependencies:
			if err := w.pollDependencies(ctx, req.AppID); err != nil {
				l.Info("unable to poll app dependencies: %w", zap.Error(err))
			}
		case OperationProvision:
			if err := w.provision(ctx, req.AppID, req.SandboxMode); err != nil {
				l.Info("unable to provision app: %w", zap.Error(err))
			}
		case OperationReprovision:
			if err := w.reprovision(ctx, req.AppID, req.SandboxMode); err != nil {
				l.Info("unable to reprovision app: %w", zap.Error(err))
			}
		case OperationUpdateSandbox:
			if err := w.updateSandbox(ctx, req.AppID, signal.SandboxReleaseID, req.SandboxMode); err != nil {
				l.Info("unable to provision app: %w", zap.Error(err))
			}
		case OperationDeprovision:
			if err := w.deprovision(ctx, req.AppID, req.SandboxMode); err != nil {
				l.Info("unable to deprovision app: %w", zap.Error(err))
				return
			}
			finished = true
		}
	})
	for !finished {
		selector.Select(ctx)
	}

	return nil
}
