package restart

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	runnergracefulshutdown "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/gracefulshutdown"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "org-restart"

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

	runners, err := activities.AwaitGetRunnersByID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get runners: %w", err)
	}

	for _, runner := range runners {
		if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   runner.ID,
			OwnerType: "runners",
			Signal:    &runnergracefulshutdown.Signal{RunnerID: runner.ID},
		}); err != nil {
			l.Error("unable to enqueue graceful shutdown signal", zap.String("runner_id", runner.ID), zap.Error(err))
			return fmt.Errorf("unable to enqueue graceful shutdown signal for runner %s: %w", runner.ID, err)
		}
	}

	return nil
}
