package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	EventLoopWorkflowName string = "AppEventLoop"
)

func EventLoopWorkflowID(orgID string) string {
	return fmt.Sprintf("%s-event-loop", orgID)
}

func (w *Workflows) AppEventLoop(ctx workflow.Context, appID string) error {
	l := zap.L()

	signalChan := workflow.GetSignalChannel(ctx, appID)
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
		case OperationProvision:
			if err := w.provision(ctx, appID, signal.DryRun); err != nil {
				l.Info("unable to provision app: %w", zap.Error(err))
			}
		case OperationUpdateSandbox:
			if err := w.updateSandbox(ctx, appID, signal.SandboxReleaseID, signal.DryRun); err != nil {
				l.Info("unable to provision app: %w", zap.Error(err))
			}
		case OperationDeprovision:
			if err := w.deprovision(ctx, appID, signal.DryRun); err != nil {
				l.Info("unable to deprovision app: %w", zap.Error(err))
			}
		}
	})
	for {
		selector.Select(ctx)
	}

	return nil
}
