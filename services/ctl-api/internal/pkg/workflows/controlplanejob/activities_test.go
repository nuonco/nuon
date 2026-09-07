package controlplanejob

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestExecutionStatusForError(t *testing.T) {
	tests := map[string]struct {
		err    error
		status app.RunnerJobExecutionStatus
	}{
		"failure": {
			err:    errors.New("build failed"),
			status: app.RunnerJobExecutionStatusFailed,
		},
		"timeout": {
			err:    temporal.NewTimeoutError(enumspb.TIMEOUT_TYPE_START_TO_CLOSE, errors.New("deadline exceeded")),
			status: app.RunnerJobExecutionStatusTimedOut,
		},
		"cancellation": {
			err:    temporal.NewCanceledError(),
			status: app.RunnerJobExecutionStatusCancelled,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.status, executionStatusForError(test.err))
		})
	}
}

func TestRedactedStringMapToHstore(t *testing.T) {
	const secret = "control-plane-secret"
	raw := "https://api.example.com/state?token=" + secret + "&operation=plan"
	meta := map[string]string{
		"error_output": raw,
		"handler":      "control-plane",
	}

	got := redactedStringMapToHstore(meta)
	require.Equal(t, raw, meta["error_output"])
	require.NotNil(t, got["error_output"])
	require.NotContains(t, *got["error_output"], secret)
	require.Contains(t, *got["error_output"], "token=[REDACTED]")
	require.Equal(t, "control-plane", *got["handler"])
}
