package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// IsRetryableResponse is the response from the "is-retryable" update handler.
type IsRetryableResponse struct {
	Retryable      bool   `json:"retryable"`
	Skippable      bool   `json:"skippable"`
	AutoRetry      bool   `json:"auto_retry"`
	MaxRetries     int    `json:"max_retries"`
	MaxAutoRetries int    `json:"max_auto_retries"`
	RetryGroup     bool   `json:"retry_group"`
	RetryIndex     int    `json:"retry_index"`
	StepID         string `json:"step_id"`
}

func (s *Signal) isRetryableHandler(ctx workflow.Context) (*IsRetryableResponse, error) {
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return nil, err
	}

	resp := &IsRetryableResponse{
		Retryable:  step.Retryable,
		Skippable:  step.Skippable,
		MaxRetries: signal.DefaultMaxRetries,
		RetryIndex: step.RetryIndex,
		StepID:     step.ID,
	}

	// Read capabilities from the inner signal's interfaces.
	sig := stepSignal(step)
	if sig != nil {
		if ar, ok := sig.(signal.SignalWithAutoRetry); ok {
			resp.AutoRetry = ar.AutoRetry()
		}
		if mr, ok := sig.(signal.SignalWithMaxRetries); ok {
			resp.MaxRetries = mr.MaxRetries()
		}
		resp.MaxAutoRetries = resp.MaxRetries
		if mar, ok := sig.(signal.SignalWithMaxAutoRetries); ok {
			resp.MaxAutoRetries = mar.MaxAutoRetries(ctx)
		}
		if rg, ok := sig.(signal.SignalWithRetryGroup); ok {
			resp.RetryGroup = rg.RetryGroup()
		}
	}

	return resp, nil
}
