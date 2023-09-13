package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	EventLoopWorkflowName string = "ComponentEventLoop"
)

func EventLoopWorkflowID(componentID string) string {
	return fmt.Sprintf("%s-event-loop", componentID)
}

func (w *Workflows) ComponentEventLoop(ctx workflow.Context, appID string) error {
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
		case OperationPollDependencies:
			if err := w.pollDependencies(ctx, appID); err != nil {
				l.Info("unable to poll app status for readiness: %w", zap.Error(err))
			}
		case OperationBuild:
			if err := w.build(ctx, appID, signal.BuildID, signal.DryRun); err != nil {
				l.Info("unable to build component: %w", zap.Error(err))
			}
		case OperationDelete:
			if err := w.delete(ctx, appID, signal.DryRun); err != nil {
				l.Info("unable to delete component: %w", zap.Error(err))
			}
		}
	})
	for {
		selector.Select(ctx)
	}
}
