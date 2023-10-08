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

type ComponentEventLoopRequest struct {
	ComponentID string
	SandboxMode bool
}

func (w *Workflows) ComponentEventLoop(ctx workflow.Context, req ComponentEventLoopRequest) error {
	l := zap.L()

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, req.ComponentID)
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
			if err := w.pollDependencies(ctx, req.ComponentID); err != nil {
				l.Info("unable to poll app status for readiness: %w", zap.Error(err))
			}
		case OperationBuild:
			if err := w.build(ctx, req.ComponentID, signal.BuildID, req.SandboxMode); err != nil {
				l.Info("unable to build component: %w", zap.Error(err))
			}
		case OperationDelete:
			if err := w.delete(ctx, req.ComponentID, req.SandboxMode); err != nil {
				l.Info("unable to delete component: %w", zap.Error(err))
			}
			finished = true
		}
	})
	for !finished {
		selector.Select(ctx)
	}

	return nil
}
