package restart

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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

	evs, err := activities.AwaitGetEventLoopsByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get event loops: %w", err)
	}

	for _, ev := range evs {
		// skip restarting own namespace to avoid infinite loop
		if ev.Namespace == "orgs" {
			continue
		}

		if err := queueclient.AwaitRestart(ctx, ev.ID); err != nil {
			l.Error("unable to restart queue", zap.String("queue_id", ev.ID), zap.String("namespace", ev.Namespace))
			return fmt.Errorf("unable to restart queue %s: %w", ev.ID, err)
		}
	}

	return nil
}
