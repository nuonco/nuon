package signal

import (
	"encoding/json"
	"time"

	"go.temporal.io/sdk/workflow"
)

type SignalType string

type Signal interface {
	Type() SignalType

	Validate(ctx workflow.Context) error
	Execute(ctx workflow.Context) error
}

type SleepAfter interface {
	SleepAfter() time.Duration
}

type AutoExecuteOnTerminalStart interface {
	AutoExecuteOnTerminalStart() bool
	AutoExecuteReady() bool
	AutoExecuteDeclined() bool
}

type CompletionCallbacksWorkflow interface {
	CompletionCallbacksWorkflowID() string
}

const DefaultSleepAfter = 1 * time.Minute

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
