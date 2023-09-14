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

func (w *Workflows) AppEventLoop(ctx workflow.Context, appID string) error {
	l := zap.L()

	finished := false
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
		case OperationPollDependencies:
			if err := w.pollDependencies(ctx, appID); err != nil {
				l.Info("unable to poll app dependencies: %w", zap.Error(err))
			}
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
			finished = true
		}
	})
	for !finished {
		selector.Select(ctx)
	}

	return nil
}
