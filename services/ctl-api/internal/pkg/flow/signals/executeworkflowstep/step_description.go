package executeworkflowstep

import (
	stderrors "errors"

	"go.temporal.io/sdk/temporal"
)

// stepHumanDescription returns a user-facing error message for a failed step.
// Duplicated from flow.StepHumanDescription to avoid import cycle.
func stepHumanDescription(err error) string {
	var appErr *temporal.ApplicationError
	if stderrors.As(err, &appErr) && appErr.NonRetryable() {
		return appErr.Message()
	}
	return "Step failed"
}
