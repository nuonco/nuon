package force_sandbox_mode

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org-force-sandbox-mode"

type Signal struct {
	signal.Hooks
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

	org, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	if org.SandboxMode {
		l.Debug("skipping, org is in sandbox mode", zap.String("name", org.Name))
	}

	if err := activities.AwaitForceSandboxModeByOrgID(ctx, s.OrgID); err != nil {
		return fmt.Errorf("unable to force sandbox mode for org: %w", err)
	}

	if err := activities.AwaitForceRunnersSandboxModeByOrgID(ctx, s.OrgID); err != nil {
		return fmt.Errorf("unable to force sandbox mode for runners for org: %w", err)
	}

	return nil
}
