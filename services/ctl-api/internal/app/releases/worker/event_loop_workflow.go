package worker

import (
	"fmt"

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
	l := zap.L()

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

		switch signal.Operation {
		case OperationPollDependencies:
			if err := w.pollDependencies(ctx, req.ReleaseID); err != nil {
				l.Info("unable to poll dependencies: %w", zap.Error(err))
			}
		case OperationProvision:
			if err := w.provision(ctx, req.ReleaseID, req.SandboxMode); err != nil {
				l.Info("unable to provision release: %w", zap.Error(err))
			}
			finished = true
		}
	})
	for !finished {
		selector.Select(ctx)
	}

	return nil
}
