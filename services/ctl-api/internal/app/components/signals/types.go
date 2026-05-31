package signals

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	TemporalNamespace string = "components"
)

// SignalType is a string identifier for signal operations.
type SignalType = string

const (
	OperationCreated             SignalType = "created"
	OperationRestart             SignalType = "restart"
	OperationBuild               SignalType = "build"
	OperationQueueBuild          SignalType = "queue_build"
	OperationProvision           SignalType = "provision"
	OperationDelete              SignalType = "delete"
	OperationPollDependencies    SignalType = "poll_dependencies"
	OperationConfigCreated       SignalType = "config_created"
	OperationUpdateComponentType SignalType = "update_component_type"
)

// Signal contains the details of a component signal operation.
type Signal struct {
	Type string `validate:"required"`

	BuildID       string            `validate:"required_if=Operation build"`
	ComponentType app.ComponentType `validate:"required_if=Operation update_component_type"`
}

// EventLoopRequest holds the core request fields previously provided by the eventloop package.
type EventLoopRequest struct {
	ID          string
	SandboxMode bool
}

// RequestSignal is the parameter type for component workflow functions.
type RequestSignal struct {
	*Signal
	EventLoopRequest
}

// NewRequestSignal constructs a RequestSignal from an EventLoopRequest and a Signal.
func NewRequestSignal(req EventLoopRequest, signal *Signal) RequestSignal {
	return RequestSignal{
		Signal:           signal,
		EventLoopRequest: req,
	}
}
