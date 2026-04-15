package restart

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "install-restart"

type Signal struct {
	signal.Hooks
	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithInit = (*Signal)(nil)

func (s *Signal) Init(_ workflow.Context) error {
	s.Hooks.InstallID = &s.InstallID
	s.Hooks.Operation = "install-restart"
	return nil
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}

	// Validate install exists
	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "install not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Restart signal - currently a no-op (copied from worker/restarted.go)
	// The restart logic is handled by the event loop infrastructure
	return nil
}
