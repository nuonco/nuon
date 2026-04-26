package restart_queues

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org-restart-queues"

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

	for _, queue := range queues {
		if err := queueclient.AwaitRestart(ctx, queue.ID); err != nil {
			l.Error("unable to restart queue", zap.String("queue_id", queue.ID))
			return fmt.Errorf("unable to restart queue %s: %w", queue.ID, err)
		}

		emitters, err := emitterclient.AwaitGetEmittersByQueueID(ctx, queue.ID)
		if err != nil {
			l.Error("unable to get emitters for queue", zap.String("queue_id", queue.ID))
			return fmt.Errorf("unable to get emitters for queue %s: %w", queue.ID, err)
		}

		for _, emitter := range emitters {
			if _, err := emitterclient.AwaitRestartEmitterWorkflow(ctx, emitter.ID); err != nil {
				l.Error("unable to restart emitter",
					zap.String("emitter_id", emitter.ID),
					zap.String("queue_id", queue.ID))
				return fmt.Errorf("unable to restart emitter %s: %w", emitter.ID, err)
			}
		}
	}

	return nil
}
