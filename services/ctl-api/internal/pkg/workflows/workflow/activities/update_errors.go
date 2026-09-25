package activities

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.temporal.io/sdk/temporal"
)

// updateOutcomeError classifies an error returned while waiting on a workflow
// update result. A rejection from the update handler is a domain outcome (e.g.
// "max retries exhausted", "step is not retryable") — retrying the activity
// re-sends the same update and can never succeed, so it must be non-retryable
// or the caller hangs behind Temporal's activity retry loop. Timeouts and
// cancellations are transport failures and stay retryable.
//
// "unknown update ... KnownUpdates=[...]" rejections are also transient, not
// domain outcomes: during a cold rewarm UpdateWithStart starts the handler
// workflow and the update can land before the workflow registers its update
// handlers. Retrying the activity re-sends the update after registration.
func updateOutcomeError(msg string, err error) error {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if strings.Contains(err.Error(), "unknown update") {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return temporal.NewNonRetryableApplicationError(
		fmt.Sprintf("%s: %s", msg, err.Error()), "update-rejected", err)
}
