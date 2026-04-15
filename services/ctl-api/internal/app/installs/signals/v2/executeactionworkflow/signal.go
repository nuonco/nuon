package executeactionworkflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "execute-action-workflow"

// Signal is an alias for actionworkflowrun.Signal
// The execute-action-workflow operation uses the same logic as action-workflow-run
type Signal struct {
	*actionworkflowrun.Signal
}

var _ signal.Signal = &Signal{}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.Signal == nil {
		s.Signal = &actionworkflowrun.Signal{}
	}
	return s.Signal.Validate(ctx)
}

func (s *Signal) Execute(ctx workflow.Context) error {
	if s.Signal == nil {
		s.Signal = &actionworkflowrun.Signal{}
	}
	return s.Signal.Execute(ctx)
}
