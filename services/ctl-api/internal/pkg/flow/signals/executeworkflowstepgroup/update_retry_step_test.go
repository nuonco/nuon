package executeworkflowstepgroup

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
)

func TestForwardRetryStepErrorPreservesNonRetryable(t *testing.T) {
	cause := temporal.NewNonRetryableApplicationError(
		"max retries exhausted (10/10)",
		"MAX_RETRIES_EXHAUSTED",
		nil,
	)
	err := forwardRetryStepError("stp-example", fmt.Errorf("activity failed: %w", cause))

	appErr, ok := err.(*temporal.ApplicationError)
	require.True(t, ok)
	require.True(t, appErr.NonRetryable())
	require.Equal(t, "MAX_RETRIES_EXHAUSTED", appErr.Type())
}

func TestForwardRetryStepErrorLeavesTransientErrorsRetryable(t *testing.T) {
	err := forwardRetryStepError("stp-example", fmt.Errorf("connection reset"))

	_, ok := err.(*temporal.ApplicationError)
	require.False(t, ok)
	require.ErrorContains(t, err, "unable to forward retry to step stp-example")
}
