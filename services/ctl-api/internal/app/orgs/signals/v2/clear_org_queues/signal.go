package clearorgqueues

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	orgactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org-clear-queues"

type Signal struct {
	OrgID string `json:"org_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)
	l.Info("clearing all queues for org", zap.String("org_id", s.OrgID))

	opts := &workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	resp, err := orgactivities.AwaitClearOrgQueuesByOrgID(ctx, s.OrgID, opts)
	if err != nil {
		return fmt.Errorf("unable to clear org queues: %w", err)
	}

	l.Info("org queue clear complete",
		zap.Int("queues_cleared", resp.QueuesCleared),
		zap.Int("signals_cancelled", resp.SignalsCancelled))

	return nil
}
