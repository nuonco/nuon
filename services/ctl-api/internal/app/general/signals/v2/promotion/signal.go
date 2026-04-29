package promotion

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	generalactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "general-promotion"

var _ signal.Signal = (*Signal)(nil)

type Signal struct {
	Tag string `json:"tag"`
}

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(_ workflow.Context) error {
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l, _ := log.WorkflowLogger(ctx)

	// Mark all active runner processes for shutdown in a single query.
	processResp, err := generalactivities.AwaitMarkActiveProcessesForShutdown(ctx, generalactivities.MarkActiveProcessesForShutdownRequest{})
	if err != nil {
		return fmt.Errorf("unable to mark processes for shutdown: %w", err)
	}
	if l != nil {
		l.Info("marked processes for shutdown", zap.Int64("rows_affected", processResp.RowsAffected))
	}

	// Request continue-as-new on all queues so they restart with the new
	// code version.
	queueResp, err := queueclient.AwaitRequestCANAll(ctx, &queueclient.RequestCANAllRequest{})
	if err != nil {
		return fmt.Errorf("unable to request CAN on all queues: %w", err)
	}
	if l != nil {
		l.Info("requested CAN on queues", zap.Int64("rows_affected", queueResp.RowsAffected))
	}

	return nil
}
