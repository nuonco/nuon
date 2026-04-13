package flowutil

import (
	stderrors "errors"

	"go.temporal.io/sdk/temporal"
)

// StepHumanDescription returns a user-facing error message for a failed step.
// For non-retryable application errors (validation failures, user errors), it
// extracts the root cause so the user sees actionable details.
// For transient/internal errors, it returns a generic message to avoid leaking internals.
func StepHumanDescription(err error) string {
	var appErr *temporal.ApplicationError
	if stderrors.As(err, &appErr) && appErr.NonRetryable() {
		root := err
		for {
			unwrapped := stderrors.Unwrap(root)
			if unwrapped == nil {
				break
			}
			root = unwrapped
		}
		return root.Error()
	}
	return "Step failed"
}
