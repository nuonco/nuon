package worker

import (
	"errors"
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
	l := workflow.GetLogger(ctx)

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

		var err error
		switch signal.Operation {
		case OperationPollDependencies:
			err = w.pollDependencies(ctx, installID)
			if err != nil {
				l.Error("unable to poll dependencies", zap.Error(err))
				return
			}
		case OperationProvision:
			err = w.provision(ctx, installID, signal.DryRun)
			if err != nil {
				l.Error("unable to provision", zap.Error(err))
				return
			}
		case OperationReprovision:
			err = w.reprovision(ctx, installID, signal.DryRun)
			if err != nil {
				l.Error("unable to reprovision", zap.Error(err))
				return
			}
		case OperationDeprovision:
			err = w.deprovision(ctx, installID, signal.DryRun)
			if err != nil {
				l.Error("unable to deprovision", zap.Error(err))
				return
			}
			finished = true
		case OperationForgotten:
			finished = true
		case OperationDeploy:
			err = w.deploy(ctx, installID, signal.DeployID, signal.DryRun)
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
