package provision

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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
	// This is a noop signal - provision workflow does nothing
	return nil
}
