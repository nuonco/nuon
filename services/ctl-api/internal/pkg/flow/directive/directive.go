package directive

import "github.com/nuonco/nuon/services/ctl-api/internal/app"

type Step string

const (
	StepContinue Step = "continue"

	StepStop Step = "stop"

	StepRetry Step = "retry"

	StepRetryGroup Step = "retry-group"

	StepSkipGroup Step = "skip-group"

	StepAwaitApproval Step = "await-approval"

	StepAwaitRetry Step = "await-retry"
)

func (d Step) IsTerminal() bool {
	switch d {
	case StepContinue, StepStop, StepRetry, StepRetryGroup, StepSkipGroup:
		return true
	default:
		return false
	}
}

type StepResult struct {
	Directive Step

	Reason string

	SiblingStatus app.Status

	FutureStatus app.Status
}

func NewStepResult(d Step) StepResult {
	return StepResult{
		Directive:     d,
		SiblingStatus: app.StatusDiscarded,
		FutureStatus:  app.StatusNotAttempted,
	}
}

type Group string

const (
	GroupContinue Group = "continue"

	GroupStop Group = "stop"

	GroupRetryGroup Group = "retry-group"

	GroupSkipGroup Group = "skip-group"

	GroupAwaitApproval Group = "await-approval"

	GroupAwaitRetry Group = "await-retry"
)

const MetadataKey = "directive"
