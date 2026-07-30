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
// how long the handler workflow stays alive after execution. This grace period
// allows subsequent signals to reuse the running workflow via update-with-start
// instead of starting a new one. Defaults to 1 minute if not implemented.
// Return 0 to terminate the handler workflow immediately after execution.
type SleepAfter interface {
	SleepAfter() time.Duration
}

// AutoExecuteOnTerminalStart is an optional interface for resident signals
// whose Handler must re-enter Execute when it is (re)started by an
// update-with-start after the signal already completed successfully.
//
// A resident signal's Execute() parks instead of returning, then idles out and
// closes the Handler with the QueueSignal marked StatusSuccess. A later
// update-with-start (e.g. append-step / retry-step) starts a fresh Handler run,
// but the queue dispatcher never re-drives ready/validate/execute on a terminal
// signal — so without this hook the conductor loop never restarts and the
// appended/retried work is never consumed. Returning true lets the Handler
// self-drive validate→execute exactly once on such a re-warm.
// AutoExecuteReady reports that a mutating update (retry/append/resume/skip)
// has started, so the Handler should re-drive validate→execute.
// AutoExecuteDeclined reports that only read-only updates (polls,
// retryability checks) have run and none are in flight, so the Handler should
// finish without re-executing the terminal signal.
type AutoExecuteOnTerminalStart interface {
	AutoExecuteOnTerminalStart() bool
	AutoExecuteReady() bool
	AutoExecuteDeclined() bool
}

// CompletionCallbacksWorkflow identifies resident flow signals whose parent
// callbacks must reflect the persisted workflow outcome instead of the queue
// signal's transport outcome. An empty ID preserves the default behavior.
type CompletionCallbacksWorkflow interface {
	CompletionCallbacksWorkflowID() string
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
