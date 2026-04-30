package signal

import (
	"errors"

	"go.temporal.io/sdk/temporal"
)

// HumanError extracts a clean, customer-facing error message from a Temporal
// error chain. Temporal wraps errors in layers of ActivityError,
// ChildWorkflowExecutionError, and ApplicationError — each prepending
// internal details (workflow IDs, event IDs, identity strings) that are
// meaningless to end users.
//
// This function unwraps through those layers and returns the innermost
// ApplicationError message, which is the actual domain error set by signal
// code. If no ApplicationError is found, it falls back to the outermost
// error message.
func HumanError(err error) string {
	if err == nil {
		return ""
	}

	// Walk the error chain looking for the innermost ApplicationError,
	// which carries the actual domain message.
	var best string
	current := err
	for current != nil {
		var appErr *temporal.ApplicationError
		if errors.As(current, &appErr) {
			best = appErr.Message()
		}
		current = errors.Unwrap(current)
	}

	if best != "" {
		return best
	}

	return err.Error()
}
