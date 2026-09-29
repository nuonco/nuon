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

type Params struct {
	MW  metrics.Writer
	Cfg *internal.Config
	V   *validator.Validate

	QueueSignalID string
}

func ApplyParams(sig Signal, params *Params) {
	if sp, ok := sig.(SignalWithParams); ok {
		sp.WithParams(params)
	}
}

type SignalWithStepContext interface {
	Signal

	SetStepContext(stepID, flowID string)
}

func ApplyStepContext(sig Signal, stepID, flowID string) {
	if sc, ok := sig.(SignalWithStepContext); ok {
		sc.SetStepContext(stepID, flowID)
	}
}

func ApplyRetryCount(sig Signal, retryIndex, groupRetryIndex int) {
	if rc, ok := sig.(SignalWithRetryCount); ok {
		rc.SetRetryCount(retryIndex, groupRetryIndex)
	}
}
