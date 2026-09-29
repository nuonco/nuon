package executeworkflowstep

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
)

func resolveFailureDirective(skipOnFailure, retryGroup, skipAutoRetry bool, retryIndex, maxRetries, maxAutoRetries int) directive.Step {
	nextRetryIndex := retryIndex + 1

	if nextRetryIndex > maxRetries {
		if skipOnFailure {
			return directive.StepContinue
		}
		return directive.StepStop
	}

	if skipAutoRetry || nextRetryIndex > maxAutoRetries {
		if skipOnFailure && maxAutoRetries >= maxRetries {
			return directive.StepContinue
		}
		return directive.StepAwaitRetry
	}

	if retryGroup {
		return directive.StepRetryGroup
	}
	return directive.StepRetry
}
