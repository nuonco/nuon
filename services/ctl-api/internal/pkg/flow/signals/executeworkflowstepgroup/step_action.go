package executeworkflowstepgroup

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
)

type stepAction int

const (
	actionAdvance stepAction = iota
	actionRetryStep
	actionStopGroup
	actionRetryGroup
	actionSkipGroup
	actionAwaitApproval
	actionAwaitRetry
)

// why: resolveStepAction maps a completed step's directive to the group's next
// action. Unknown or empty directives fail closed to actionStopGroup: only an
// explicit success or retry directive may lead to another step dispatch, so a
// step whose domain outcome cannot be read never lets the group carry on.
func resolveStepAction(d directive.Step, residentFlow, manualRetry bool) stepAction {
	switch d {
	case directive.StepContinue:
		return actionAdvance

	case directive.StepRetry:
		if residentFlow && manualRetry {
			return actionAwaitRetry
		}
		return actionRetryStep

	case directive.StepStop:
		return actionStopGroup

	case directive.StepRetryGroup:
		if residentFlow && manualRetry {
			return actionAwaitRetry
		}
		return actionRetryGroup

	case directive.StepSkipGroup:
		return actionSkipGroup

	case directive.StepAwaitApproval:
		return actionAwaitApproval

	case directive.StepAwaitRetry:
		if residentFlow {
			return actionAwaitRetry
		}
		return actionAdvance

	default:
		return actionStopGroup
	}
}
