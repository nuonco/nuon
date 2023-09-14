package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	EventLoopWorkflowName string = "InstallEventLoop"
)

func EventLoopWorkflowID(installID string) string {
	return fmt.Sprintf("%s-event-loop", installID)
}

func (w *Workflows) InstallEventLoop(ctx workflow.Context, installID string) error {
	l := zap.L()

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, installID)
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
			if err := w.pollDependencies(ctx, installID); err != nil {
				l.Info("unable to poll install dependencies: %w", zap.Error(err))
			}
		case OperationProvision:
			if err := w.provision(ctx, installID, signal.DryRun); err != nil {
				l.Info("unable to provision install: %w", zap.Error(err))
			}
		case OperationDeprovision:
			if err := w.deprovision(ctx, installID, signal.DryRun); err != nil {
				l.Info("unable to deprovision install: %w", zap.Error(err))
			}
			finished = true
		case OperationDeploy:
			if err := w.deploy(ctx, installID, signal.DeployID, signal.DryRun); err != nil {
				l.Info("unable to perform deploy: %w", zap.Error(err))
			}
		}
	})
	for !finished {
		selector.Select(ctx)
	}

	return nil
}
