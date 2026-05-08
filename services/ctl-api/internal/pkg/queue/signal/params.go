package signal

import (
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type SignalWithParams interface {
	Signal

	WithParams(*Params)
}

// NOTE: we have a limited set of FX dependencies we can use in a signal, because they are built with signals and pass
// to and from workflows.
type Params struct {
	MW  metrics.Writer
	Cfg *internal.Config
	V   *validator.Validate

	// QueueSignalID is the ID of the queue signal that this signal is associated with.
	// This allows signals to reference their own queue signal record if needed.
	QueueSignalID string
}

// ApplyParams checks if the signal implements signalWithParams and calls WithParams if so.
func ApplyParams(sig Signal, params *Params) {
	if sp, ok := sig.(SignalWithParams); ok {
		sp.WithParams(params)
	}
}

// SignalWithStepContext is implemented by signals that need the workflow step ID and flow ID
// injected at step generation time, before the step is persisted to the database.
type SignalWithStepContext interface {
	Signal

	SetStepContext(stepID, flowID string)
}

// ApplyStepContext sets step context on signals that implement SignalWithStepContext.
func ApplyStepContext(sig Signal, stepID, flowID string) {
	if sc, ok := sig.(SignalWithStepContext); ok {
		sc.SetStepContext(stepID, flowID)
	}
}

// ApplyRetryCount sets the retry count on signals that implement SignalWithRetryCount.
func ApplyRetryCount(sig Signal, retryIndex, groupRetryIndex int) {
	if rc, ok := sig.(SignalWithRetryCount); ok {
		rc.SetRetryCount(retryIndex, groupRetryIndex)
	}
}
