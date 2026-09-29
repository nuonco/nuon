package executeworkflowstepgroup

import (
	stderrors "errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type RetryStepRequest struct {
	StepID string `json:"step_id"`
}

type RetryStepResponse struct {
	Retryable bool   `json:"retryable"`
	Directive string `json:"directive"`
}

func (s *Signal) retryStepHandler(ctx workflow.Context, req RetryStepRequest) (*RetryStepResponse, error) {
	resp, err := activities.AwaitForwardCreateStepRetry(ctx, activities.ForwardCreateStepRetryRequest{
		StepID: req.StepID,
	})
	if err != nil {
		return nil, forwardRetryStepError(req.StepID, err)
	}

	return &RetryStepResponse{
		Retryable: true,
		Directive: resp.Directive,
	}, nil
}

func forwardRetryStepError(stepID string, err error) error {
	message := fmt.Sprintf("unable to forward retry to step %s", stepID)
	var appErr *temporal.ApplicationError
	if stderrors.As(err, &appErr) && appErr.NonRetryable() {
		return temporal.NewNonRetryableApplicationError(message, appErr.Type(), err)
	}
	return fmt.Errorf("%s: %w", message, err)
}
