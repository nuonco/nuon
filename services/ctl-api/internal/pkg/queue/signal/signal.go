package signal

import (
	"encoding/json"
	"time"

	"go.temporal.io/sdk/workflow"
)

type SignalType string

type Signal interface {
	Type() SignalType
	GetHooks() *Hooks

	// workflow handler methods
	Validate(ctx workflow.Context) error
	Execute(ctx workflow.Context) error
}

// SleepAfter is an optional interface that signals can implement to control
// how long the handler workflow stays alive after execution. This grace period
// allows subsequent signals to reuse the running workflow via update-with-start
// instead of starting a new one. Defaults to 1 minute if not implemented.
// Return 0 to terminate the handler workflow immediately after execution.
type SleepAfter interface {
	SleepAfter() time.Duration
}

const DefaultSleepAfter = 1 * time.Minute

// Raw is a signal envelope for enqueueing without importing the concrete signal
// package. The queue handler deserializes into the real registered type via the
// catalog at execution time.
type Raw struct {
	Hooks
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
