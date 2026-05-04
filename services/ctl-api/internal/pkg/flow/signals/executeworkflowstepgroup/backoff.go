package executeworkflowstepgroup

import (
	"math/rand/v2"
	"time"

	"go.temporal.io/sdk/workflow"
)

// backoffSchedule defines the base auto-retry delay for each retry attempt.
// retryIndex is 1-indexed: index 1 = first retry uses backoffSchedule[0], etc.
// Sized for the default of 3 auto-retries; further retries reuse the cap.
var backoffSchedule = []time.Duration{
	30 * time.Second,
	2 * time.Minute,
	5 * time.Minute,
}

// backoffMaxDelay caps the post-jitter delay for any retry, even when
// callers configure higher MaxAutoRetries.
const backoffMaxDelay = 10 * time.Minute

// computeAutoRetryBackoff returns the wait duration before the Nth auto-retry
// executes. retryIndex is 1-indexed (1 = first auto-retry). Returns 0 when
// retryIndex < 1.
//
// Applies equal-jitter: half of base + uniform[0, base/2). Random source is
// pulled through workflow.SideEffect so the value is durable across replays.
// Result is capped at backoffMaxDelay.
func computeAutoRetryBackoff(ctx workflow.Context, retryIndex int) time.Duration {
	if retryIndex < 1 {
		return 0
	}

	idx := retryIndex - 1
	if idx >= len(backoffSchedule) {
		idx = len(backoffSchedule) - 1
	}
	base := backoffSchedule[idx]

	half := base / 2
	var jitterNs int64
	if half > 0 {
		_ = workflow.SideEffect(ctx, func(workflow.Context) interface{} {
			return rand.Int64N(int64(half))
		}).Get(&jitterNs)
	}

	delay := half + time.Duration(jitterNs)
	if delay > backoffMaxDelay {
		delay = backoffMaxDelay
	}
	return delay
}
