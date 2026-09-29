package flow

import (
	stderrors "errors"

	"go.temporal.io/sdk/temporal"
)

const DirectiveKey = "directive"

const (
	DirectiveContinue = "continue"
	DirectiveStop     = "stop"

	DirectiveAwaitApproval = "await-approval"

	DirectiveSkipGroup  = "skip-group"
	DirectiveRetry      = "retry"
	DirectiveRetryGroup = "retry-group"
)

func StepHumanDescription(err error) string {
	var appErr *temporal.ApplicationError
	if stderrors.As(err, &appErr) && appErr.NonRetryable() {
		return appErr.Message()
	}
	return "Step failed"
}
