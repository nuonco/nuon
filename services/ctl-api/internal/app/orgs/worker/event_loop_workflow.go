package worker

import (
	"errors"
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

type OrgEventLoopRequest struct {
	OrgID       string
	SandboxMode bool
}

func (w *Workflows) OrgEventLoop(ctx workflow.Context, req OrgEventLoopRequest) error {
	l := zap.L()
	tags := w.defaultTags(req.OrgID, req.SandboxMode)

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, req.OrgID)
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
		// OperationProvision
		case OperationProvision:
			err := w.provision(ctx, req.OrgID, req.SandboxMode)
			w.writeStatusMetric(ctx, "signal", err, tags, "signal", "provision")

		// OperationReprovision
		case OperationReprovision:
			err := w.reprovision(ctx, req.OrgID, req.SandboxMode)
			w.writeStatusMetric(ctx, "signal", err, tags, "signal", "reprovision")

		// OperationDeprovision
		case OperationDeprovision:
			err := w.deprovision(ctx, req.OrgID, req.SandboxMode)
			w.writeStatusMetric(ctx, "signal", err, tags, "signal", "deprovision")

		// OperationRestart
		case OperationRestart:
			w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
				OrgID:       req.OrgID,
				SandboxMode: req.SandboxMode,
			})
			w.writeStatusMetric(ctx, "signal", nil, tags, "signal", "restart")

		// OperationDelete
		case OperationDelete:
			err := w.delete(ctx, req.OrgID, req.SandboxMode)
			w.writeStatusMetric(ctx, "signal", err, tags, "signal", "delete")
			if err != nil {
				return
			}

			finished = true
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
