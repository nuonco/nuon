package signal

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type SignalType string

type Signal interface {
	Type() SignalType

	// workflow handler methods
	Validate(ctx workflow.Context) error
	Execute(ctx workflow.Context) error
}

// SleepAfter is an optional interface that signals can implement to control
// how long the handler sleeps after execution. Defaults to 1 minute if not implemented.
// Return 0 or any duration < 1 second to skip the sleep entirely.
type SleepAfter interface {
	SleepAfter() time.Duration
}

const DefaultSleepAfter = 1 * time.Minute
