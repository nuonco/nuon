package flow

import (
	stderrors "errors"

	"go.temporal.io/sdk/temporal"
)

// DirectiveKey is the metadata key used to communicate step execution directives
// from the step workflow back to the parent conductor.
const DirectiveKey = "directive"

// Step directives tell the parent conductor how to proceed after a step completes.
const (
	DirectiveContinue = "continue"
	DirectiveStop     = "stop"

	DirectiveAwaitApproval = "await-approval"

	DirectiveSkipGroup  = "skip-group"
	DirectiveRetry      = "retry"
	DirectiveRetryGroup = "retry-group"
)

// StepHumanDescription returns a user-facing error message for a failed step.
// For non-retryable application errors (validation failures, user errors), it
// extracts the root cause so the user sees actionable details.
// For transient/internal errors, it returns a generic message to avoid leaking internals.
func StepHumanDescription(err error) string {
	var appErr *temporal.ApplicationError
	if stderrors.As(err, &appErr) && appErr.NonRetryable() {
		return appErr.Message()
	}
	return "Step failed"
}
