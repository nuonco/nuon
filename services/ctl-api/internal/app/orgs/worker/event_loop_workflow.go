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
		case OperationProvision:
			if err := w.provision(ctx, req.OrgID, req.SandboxMode); err != nil {
				l.Info("unable to provision org: %w", zap.Error(err))
			}
		case OperationReprovision:
			if err := w.reprovision(ctx, req.OrgID, req.SandboxMode); err != nil {
				l.Info("unable to reprovision org: %w", zap.Error(err))
			}
		case OperationDeprovision:
			if err := w.deprovision(ctx, req.OrgID, req.SandboxMode); err != nil {
				l.Info("unable to deprovision org: %w", zap.Error(err))
			}
		case OperationRestart:
			w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
				OrgID:       req.OrgID,
				SandboxMode: req.SandboxMode,
			})
		case OperationDelete:
			if err := w.delete(ctx, req.OrgID, req.SandboxMode); err != nil {
				l.Info("unable to delete org: %w", zap.Error(err))
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
