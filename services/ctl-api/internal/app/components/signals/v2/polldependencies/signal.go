package polldependencies

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

type Signal struct {
	signal.Hooks
	ComponentID string `json:"component_id" validate:"required"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.ComponentID == "" {
		return errors.New("component_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// TODO(ja): Components don't have a status field, so we can't update them if this fails.
	// Not sure if that's a problem or not.
	for {
		currentApp, err := activities.AwaitGetComponentAppByComponentID(ctx, s.ComponentID)
		if err != nil {
			return fmt.Errorf("unable to get component app: %w", err)
		}

		if currentApp.Status == "active" {
			return nil
		}
		if currentApp.Status == "error" {
			return fmt.Errorf("app failed: %s", currentApp.Org.StatusDescription)
		}

		workflow.Sleep(ctx, defaultPollTimeout)
	}
}
