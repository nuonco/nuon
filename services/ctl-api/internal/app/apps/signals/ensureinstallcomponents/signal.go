package ensureinstallcomponents

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "ensure-install-components"

type Signal struct {
	AppID string `json:"app_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.AppID == "" {
		return errors.New("app_id is required")
	}

	return nil
}

// Execute fans the app's components out across every install for the app.
func (s *Signal) Execute(ctx workflow.Context) error {
	return activities.AwaitEnsureInstallComponentsByAppID(ctx, s.AppID)
}
