package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"
)

// RetryNowRequest is the input for the "retry-now" update handler.
type RetryNowRequest struct{}

// RetryNowResponse is the response from the "retry-now" update handler.
type RetryNowResponse struct{}

// retryNowHandler is called when the user clicks "Retry Now" on a step that
// is currently waiting out its auto-retry backoff. It flips s.retryNowRequested
// so the gate in Execute() breaks out of the timer wait and proceeds with the
// inner signal immediately.
//
// Safe to call when the step isn't in a waiting state — the flag is only
// observed by the wait gate in Execute().
func (s *Signal) retryNowHandler(ctx workflow.Context, req RetryNowRequest) (*RetryNowResponse, error) {
	s.retryNowRequested = true
	return &RetryNowResponse{}, nil
}
