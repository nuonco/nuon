package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	tclient "go.temporal.io/sdk/client"
)

func TestUpdateWithStartUntilAcceptedRetriesOnlyPreAcceptanceRaces(t *testing.T) {
	tests := map[string]struct {
		errors   []error
		attempts int
	}{
		"exact unknown update": {
			errors:   []error{errors.New("unknown update skip-step. KnownUpdates=[]"), nil},
			attempts: 2,
		},
		"closing workflow": {
			errors:   []error{errors.New("update was aborted by closing workflow"), nil},
			attempts: 2,
		},
		"different unknown update": {
			errors:   []error{errors.New("unknown update retry-step. KnownUpdates=[skip-step]")},
			attempts: 1,
		},
		"completed before update": {
			errors:   []error{errors.New("Workflow Update failed because the Workflow completed before the Update completed")},
			attempts: 1,
		},
		"application failure": {
			errors:   []error{errors.New("skip is not allowed")},
			attempts: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			attempt := 0
			_, _ = updateWithStartUntilAccepted(t.Context(), "skip-step", func() (tclient.WorkflowUpdateHandle, error) {
				err := tc.errors[attempt]
				attempt++
				return nil, err
			})
			require.Equal(t, tc.attempts, attempt)
		})
	}
}

func TestUpdateWithStartUntilAcceptedStopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	attempts := 0
	_, err := updateWithStartUntilAccepted(ctx, "skip-step", func() (tclient.WorkflowUpdateHandle, error) {
		attempts++
		return nil, errors.New("unknown update skip-step. KnownUpdates=[]")
	})
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, attempts)
}

func TestColdHostUpdateRejectionMustBeUnwrappedAndExact(t *testing.T) {
	require.True(t, isColdHostUpdateRejection(errors.New("unknown update skip-step. KnownUpdates=[]"), "skip-step"))
	require.True(t, isColdHostUpdateRejection(errors.New("update was aborted by closing workflow"), "skip-step"))
	require.False(t, isColdHostUpdateRejection(errors.New("unknown update retry-step. KnownUpdates=[skip-step]"), "skip-step"))
	require.False(t, isColdHostUpdateRejection(errors.New("unable to forward skip-step: unknown update skip-step. KnownUpdates=[]"), "skip-step"))
	require.False(t, isColdHostUpdateRejection(errors.New("Workflow Update failed because the Workflow completed before the Update completed"), "skip-step"))
}
