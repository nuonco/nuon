package executeworkflowstep

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
)

// resolveFailureDirective decides how a failed auto-retryable step surfaces to
// the group, given its retry budget and skippability.
//
// Product invariant (property-tested): a failed step may only yield
// StepContinue when it is explicitly skippable — a non-skippable failure never
// lets the group advance past it.
func resolveFailureDirective(skippable, retryGroup, skipAutoRetry bool, retryIndex, maxRetries, maxAutoRetries int) directive.Step {
	nextRetryIndex := retryIndex + 1

	// Global ceiling — no more retries of any kind.
	if nextRetryIndex > maxRetries {
		if skippable {
			return directive.StepContinue
		}
		return directive.StepStop
	}

	// Auto-retries exhausted, or the composite error hinted that auto-retry
	// won't help. Skippable steps with no manual-retry headroom beyond the
	// auto budget continue; everything else parks for manual retry.
	if skipAutoRetry || nextRetryIndex > maxAutoRetries {
		if skippable && maxAutoRetries >= maxRetries {
			return directive.StepContinue
		}
		return directive.StepAwaitRetry
	}

	if retryGroup {
		return directive.StepRetryGroup
	}
	return directive.StepRetry
}
