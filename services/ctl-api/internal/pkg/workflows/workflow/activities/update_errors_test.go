package activities

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
)

func TestUpdateOutcomeError_DomainRejectionIsNonRetryable(t *testing.T) {
	err := updateOutcomeError("create-step-retry update failed",
		errors.New("create-step-retry rejected: max retries exhausted (2/2)"))

	var appErr *temporal.ApplicationError
	require.True(t, errors.As(err, &appErr))
	require.True(t, appErr.NonRetryable())
	require.Equal(t, "update-rejected", appErr.Type())
	require.Contains(t, appErr.Error(), "max retries exhausted")
}

func TestUpdateOutcomeError_UnknownUpdateStaysRetryable(t *testing.T) {
	err := updateOutcomeError("retry-step failed on group",
		errors.New("unknown update retry-step. KnownUpdates=[]"))

	var appErr *temporal.ApplicationError
	require.False(t, errors.As(err, &appErr),
		"unknown-update rejection must stay retryable: handlers may not be registered yet during cold rewarm")
}

func TestUpdateOutcomeError_ContextErrorsStayRetryable(t *testing.T) {
	for _, cause := range []error{context.DeadlineExceeded, context.Canceled} {
		err := updateOutcomeError("update failed", fmt.Errorf("waiting: %w", cause))

		var appErr *temporal.ApplicationError
		require.False(t, errors.As(err, &appErr),
			"context error %v must not become a non-retryable application error", cause)
		require.True(t, errors.Is(err, cause))
	}
}
