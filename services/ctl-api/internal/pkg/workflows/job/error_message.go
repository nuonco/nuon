package job

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
)

// JobErrorMessage extracts a user-facing message from an ExecuteJob error.
// The ExecuteJob workflow returns errors like "job did not succeed: <description>"
// which contain the actual runner job StatusDescription. This helper surfaces
// that message for status updates instead of using hardcoded fallback strings.
func JobErrorMessage(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Message()
	}
	msg := err.Error()
	if msg != "" {
		return msg
	}
	return fallback
}

func IsRunnerJobFailure(err error) bool {
	var appErr *temporal.ApplicationError
	return errors.As(err, &appErr) && appErr.Type() == "RUNNER_JOB_FAILED"
}
