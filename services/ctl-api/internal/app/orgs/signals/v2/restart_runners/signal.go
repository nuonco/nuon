package restart_runners

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	runnersignals "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signalsactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/signals/activities"
)

const SignalType signal.SignalType = "org-restart-runners"

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
	runners, err := activities.AwaitGetRunnersByID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get runners: %w", err)
	}

	for _, runner := range runners {
		signalsactivities.AwaitPkgSignalsSendRunnersSignal(ctx, &signalsactivities.SendSignalRequest[*runnersignals.Signal]{
			ID:     runner.ID,
			Signal: &runnersignals.Signal{Type: runnersignals.OperationGracefulShutdown},
		})
	}

	return nil
}
