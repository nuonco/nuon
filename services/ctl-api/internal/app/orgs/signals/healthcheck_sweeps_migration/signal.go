package healthcheck_sweeps_migration

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org-healthcheck-sweeps-migration"

// Signal migrates an org between per-entity healthcheck cron emitters and the
// per-org batch sweep emitters, following the org-healthcheck-sweeps flag the
// endpoint flipped before enqueueing it.
type Signal struct {
	OrgID   string `json:"org_id"`
	Enabled bool   `json:"enabled"`
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

	result, err := activities.AwaitMigrateOrgHealthcheckSweeps(ctx, activities.MigrateOrgHealthcheckSweepsRequest{
		OrgID:   s.OrgID,
		Enabled: s.Enabled,
	})
	if err != nil {
		return fmt.Errorf("unable to migrate org healthcheck sweeps: %w", err)
	}

	l.Info("org healthcheck sweeps migration complete",
		zap.String("org_id", s.OrgID),
		zap.Bool("enabled", s.Enabled),
		zap.Int("emitters_deleted", result.EmittersDeleted),
		zap.Int("emitters_created", result.EmittersCreated),
		zap.Int("queues_terminated", result.QueuesTerminated),
		zap.Int("errors", len(result.Errors)),
	)
	return nil
}
