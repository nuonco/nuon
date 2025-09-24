package signal

import "go.temporal.io/sdk/workflow"

type SignalType string

type Signal interface {
	Type() SignalType

	// workflow handler methods
	Validate(ctx workflow.Context) error
	Execute(ctx workflow.Context) error
}
