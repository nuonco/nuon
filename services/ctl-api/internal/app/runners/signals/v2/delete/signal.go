package delete

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "delete"

type Signal struct {
	signal.Hooks
	RunnerID string `json:"runner_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}

	// Validate runner exists in database
	_, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Delete the runner (soft delete via DeletedAt timestamp)
	if err := activities.AwaitDelete(ctx, activities.DeleteRequest{
		RunnerID: s.RunnerID,
	}); err != nil {
		return fmt.Errorf("unable to delete runner: %w", err)
	}

	return nil
}
