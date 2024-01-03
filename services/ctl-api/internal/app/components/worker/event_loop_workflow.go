package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/signals"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) ComponentEventLoop(ctx workflow.Context, req signals.ComponentEventLoopRequest) error {
	l := zap.L()

	finished := false

	signalChan := workflow.GetSignalChannel(ctx, req.ComponentID)
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

		switch signal.Operation {
		case signals.OperationPollDependencies:
			if err := w.pollDependencies(ctx, req.ComponentID); err != nil {
				l.Info("unable to poll app status for readiness: %w", zap.Error(err))
			}
		case signals.OperationProvision:
			if err := w.provision(ctx, req.ComponentID); err != nil {
				l.Info("unable to provision: %w", zap.Error(err))
			}
		case signals.OperationQueueBuild:
			if err := w.queueBuild(ctx, req.ComponentID); err != nil {
				l.Info("unable to handle config created: %w", zap.Error(err))
			}
		case signals.OperationBuild:
			if err := w.build(ctx, req.ComponentID, signal.BuildID, req.SandboxMode); err != nil {
				l.Info("unable to build component: %w", zap.Error(err))
			}
		case signals.OperationDelete:
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
