package polldependencies

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "poll-dependencies"

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

type Signal struct {
	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)

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
	// Poll until app is active (copied from worker/poll_dependencies.go)
	for {
		install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
		if err != nil {
			return fmt.Errorf("unable to get install and poll: %w", err)
		}

		if install.App.Status == "active" {
			return nil
		}
		workflow.Sleep(ctx, defaultPollTimeout)
	}
}
