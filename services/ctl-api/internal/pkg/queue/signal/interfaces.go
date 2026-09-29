package signal

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type CloneStepDef struct {
	Signal        Signal
	Name          string
	ExecutionType string
}

type SignalWithRetryCount interface {
	SetRetryCount(retryIndex, groupRetryIndex int)
}

type SignalWithClone interface {
	Clone(ctx workflow.Context, stepName string) ([]CloneStepDef, error)
}

type SignalWithCloneSteps = SignalWithClone

type SignalWithMaxRetries interface {
	MaxRetries() int
}

const DefaultMaxRetries = 10

type SignalWithAutoRetry interface {
	AutoRetry() bool
}

type SignalWithMaxAutoRetries interface {
	MaxAutoRetries(ctx workflow.Context) int
}

type SignalWithRetryGroup interface {
	RetryGroup() bool
}

type SignalWithInlineValidate interface {
	InlineValidate() bool
}

type SignalWithNoOpCheck interface {
	IsNoOpCheckable() bool
}

type SignalWithPolicyEvaluation interface {
	RequiresPolicyEvaluation() bool
}

type SignalWithEmptyGroupCheck interface {
	IsEmptyInstallGroup(ctx workflow.Context) (bool, error)
}

type SignalWithSkipNoops interface {
	SkipNoops(ctx workflow.Context) bool
}

type SignalWithAutoApproveOnPoliciesPassing interface {
	AutoApproveOnPoliciesPassing(ctx workflow.Context) bool
}

type SignalWithSkipCleanup interface {
	OnSkipped(ctx workflow.Context) error
}

type SignalWithOnApprove interface {
	OnApprove(ctx workflow.Context) error
}

type SignalWithOnRetry interface {
	OnRetry(ctx workflow.Context) error
}

type SignalWithOnSkip interface {
	OnSkip(ctx workflow.Context) error
}

type SignalWithOnDeny interface {
	OnDeny(ctx workflow.Context) error
}

type SignalWithApprovalValidation interface {
	ValidateApproval(ctx workflow.Context) error
}

type SignalWithSkipGroup interface {
	SkipGroup() bool
}

type SignalWithSkippable interface {
	Skippable() bool
}

func IsSkippable(sig Signal) bool {
	if s, ok := sig.(SignalWithSkippable); ok {
		return s.Skippable()
	}
	return true
}

type SignalWithFetchSteps interface {
	Signal
	SignalWithUpdateHandlers
}

type SignalWithTimeout interface {
	Timeout() time.Duration
}

type SignalWithUnboundedTimeout interface {
	UnboundedTimeout() bool
}

type SignalWithMaxInFlightAge interface {
	MaxInFlightAge() time.Duration
}

type SignalWithQueue interface {
	Queue() string
}

type SignalWithParallelizable interface {
	IsParallelizable() bool
}
