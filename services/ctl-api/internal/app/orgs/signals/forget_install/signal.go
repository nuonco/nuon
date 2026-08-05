package forgetinstall

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	orgactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org-forget-install"

type Signal struct {
	OrgID     string `json:"org_id"`
	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

// no install lookup: the row is already soft-deleted by the forget handler
func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)
	l.Info("forgetting install", zap.String("org_id", s.OrgID), zap.String("install_id", s.InstallID))

	if err := orgactivities.AwaitForgetInstallByInstallID(ctx, s.InstallID); err != nil {
		return fmt.Errorf("unable to forget install: %w", err)
	}

	l.Info("install forgotten", zap.String("install_id", s.InstallID))
	return nil
}
