package client

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/temporal"
)

func TestHandlerLostError(t *testing.T) {
	lost := []struct {
		name string
		err  error
	}{
		{"run already closed", serviceerror.NewNotFound("workflow execution already completed")},
		{"accepted then run closed", temporal.NewApplicationError("Workflow Update failed because the Workflow completed before the Update completed.", "AcceptedUpdateCompletedWorkflow")},
		{"aborted by closing run", errors.New("workflow update was aborted by closing workflow")},
	}
	for _, tc := range lost {
		t.Run(tc.name, func(t *testing.T) {
			err := handlerLostError("eager-step-groups", "qsi1", "run1", tc.err)
			require.Error(t, err)
			require.True(t, IsFetchHandlerLost(err))
			var appErr *temporal.ApplicationError
			require.True(t, errors.As(err, &appErr))
			require.True(t, appErr.NonRetryable())
			require.ErrorIs(t, err, tc.err)
		})
	}

	served := []struct {
		name string
		err  error
	}{
		{"generator failure", temporal.NewApplicationError("unable to generate steps for workflow inw1: no step generator for workflow type x", "")},
		{"plain error", errors.New("unable to decode result")},
		{"unknown update on an open run", errors.New("unknown update eager-step-groups. KnownUpdates=[cancel ready]")},
	}
	for _, tc := range served {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, handlerLostError("FetchSteps", "qsi1", "run1", tc.err))
		})
	}

	require.False(t, IsFetchHandlerLost(temporal.NewNonRetryableApplicationError("eager-step-groups update failed", "FETCH_EAGER_STEP_GROUPS_FAILED", errors.New("boom"))))
	require.False(t, IsFetchHandlerLost(nil))
}
