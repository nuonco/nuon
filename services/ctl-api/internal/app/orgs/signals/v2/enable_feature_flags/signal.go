package enable_feature_flags

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org-enable-feature-flags"

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
	if err := activities.AwaitEnableFeaturesByOrgID(ctx, s.OrgID); err != nil {
		return fmt.Errorf("unable to enable features: %w", err)
	}
	return nil
}
