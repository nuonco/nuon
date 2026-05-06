package generatestate

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "generate-state"

type Signal struct {
	InstallID string
}

var _ signal.Signal = &Signal{}
var _ signal.SignalWithAutoRetry = (*Signal)(nil)

func (s *Signal) AutoRetry() bool { return true }

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install id is required")
	}

	// Validate install exists
	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	_, err := state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       s.InstallID,
		TriggeredByID:   s.InstallID,
		TriggeredByType: "installs",
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}

	return nil
}
