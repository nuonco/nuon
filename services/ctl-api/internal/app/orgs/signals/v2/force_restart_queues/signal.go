package force_restart_queues

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org-force-restart-queues"

type Signal struct {
	OrgID string `json:"org_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	_, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("org not found: %w", err)
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	queues, err := queueclient.AwaitListQueuesByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to list queues: %w", err)
	}

	var lastErr error
	for _, queue := range queues {
		if err := queueclient.AwaitForceRestart(ctx, queue.ID); err != nil {
			l.Error("unable to force restart queue",
				zap.String("queue_id", queue.ID),
				zap.Error(err),
			)
			lastErr = err
		}
	}

	if lastErr != nil {
		return fmt.Errorf("one or more queues failed to force restart: %w", lastErr)
	}

	return nil
}
