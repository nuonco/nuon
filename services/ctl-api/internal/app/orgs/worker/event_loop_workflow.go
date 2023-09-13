package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	EventLoopWorkflowName string = "OrgEventLoop"
	defaultOrgRegion      string = "us-west-2"
)

func EventLoopWorkflowID(orgID string) string {
	return fmt.Sprintf("%s-event-loop", orgID)
}

func (w *Workflows) OrgEventLoop(ctx workflow.Context, orgID string) error {
	l := zap.L()

	signalChan := workflow.GetSignalChannel(ctx, orgID)
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
			if err := w.provision(ctx, orgID, signal.DryRun); err != nil {
				l.Info("unable to provision org: %w", zap.Error(err))
			}
		case OperationDeprovision:
			if err := w.deprovision(ctx, orgID, signal.DryRun); err != nil {
				l.Info("unable to deprovision org: %w", zap.Error(err))
			}
		}
	})
	for {
		selector.Select(ctx)
	}
}
