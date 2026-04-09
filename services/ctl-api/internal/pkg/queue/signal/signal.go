package signal

import (
	"encoding/json"
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

// Raw is a signal envelope for enqueueing without importing the concrete signal
// package. The queue handler deserializes into the real registered type via the
// catalog at execution time.
type Raw struct {
	signalType SignalType
	data       map[string]any
}

func NewRaw(typ SignalType, data map[string]any) Signal {
	return &Raw{signalType: typ, data: data}
}

func (r *Raw) Type() SignalType                  { return r.signalType }
func (r *Raw) Validate(_ workflow.Context) error { return nil }
func (r *Raw) Execute(_ workflow.Context) error  { return nil }
func (r *Raw) MarshalJSON() ([]byte, error)      { return json.Marshal(r.data) }
