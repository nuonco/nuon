package testworker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func init() {
	catalog.Register(SuccessSignalType, func() signal.Signal { return &SuccessSignal{} })
	catalog.Register(FailSignalType, func() signal.Signal { return &FailSignal{} })
	catalog.Register(SlowSignalType, func() signal.Signal { return &SlowSignal{} })
	catalog.Register(AutoRetrySignalType, func() signal.Signal { return &AutoRetrySignal{} })
	catalog.Register(RetryGroupSignalType, func() signal.Signal { return &RetryGroupSignal{} })
	catalog.Register(CountdownSignalType, func() signal.Signal { return &CountdownSignal{} })
	catalog.Register(CountdownGroupSignalType, func() signal.Signal { return &CountdownGroupSignal{} })
}

// --- SuccessSignal: completes immediately ---

const SuccessSignalType signal.SignalType = "test-flow-success"

type SuccessSignal struct{}

func (s *SuccessSignal) Type() signal.SignalType         { return SuccessSignalType }
func (s *SuccessSignal) Validate(workflow.Context) error { return nil }
func (s *SuccessSignal) Execute(workflow.Context) error  { return nil }
func (s *SuccessSignal) SleepAfter() time.Duration       { return time.Second }

// --- FailSignal: always fails ---

const FailSignalType signal.SignalType = "test-flow-fail"

type FailSignal struct {
	Reason string `json:"reason"`
}

func (s *FailSignal) Type() signal.SignalType         { return FailSignalType }
func (s *FailSignal) Validate(workflow.Context) error { return nil }
func (s *FailSignal) Execute(workflow.Context) error  { return fmt.Errorf("test failure: %s", s.Reason) }
func (s *FailSignal) SleepAfter() time.Duration       { return time.Second }

// --- SlowSignal: blocks until cancelled ---

const SlowSignalType signal.SignalType = "test-flow-slow"

type SlowSignal struct{}

func (s *SlowSignal) Type() signal.SignalType         { return SlowSignalType }
func (s *SlowSignal) Validate(workflow.Context) error { return nil }
func (s *SlowSignal) Execute(ctx workflow.Context) error {
	return workflow.Sleep(ctx, 10*time.Minute)
}

var _ signal.SignalWithCancel = (*SlowSignal)(nil)

func (s *SlowSignal) Cancel(workflow.Context) error { return nil }
func (s *SlowSignal) SleepAfter() time.Duration     { return time.Second }

// --- AutoRetrySignal: fails with auto-retry enabled, always fails ---

const AutoRetrySignalType signal.SignalType = "test-flow-auto-retry"

type AutoRetrySignal struct {
	FailUntilRetryIndex int `json:"fail_until_retry_index"`
}

func (s *AutoRetrySignal) Type() signal.SignalType         { return AutoRetrySignalType }
func (s *AutoRetrySignal) Validate(workflow.Context) error { return nil }
func (s *AutoRetrySignal) Execute(workflow.Context) error {
	return fmt.Errorf("auto-retry test failure")
}
func (s *AutoRetrySignal) AutoRetry() bool           { return true }
func (s *AutoRetrySignal) MaxRetries() int           { return 3 }
func (s *AutoRetrySignal) SleepAfter() time.Duration { return time.Second }

var _ signal.SignalWithAutoRetry = (*AutoRetrySignal)(nil)
var _ signal.SignalWithMaxRetries = (*AutoRetrySignal)(nil)

// --- RetryGroupSignal: fails and requests group retry ---

const RetryGroupSignalType signal.SignalType = "test-flow-retry-group"

type RetryGroupSignal struct{}

func (s *RetryGroupSignal) Type() signal.SignalType         { return RetryGroupSignalType }
func (s *RetryGroupSignal) Validate(workflow.Context) error { return nil }
func (s *RetryGroupSignal) Execute(workflow.Context) error {
	return fmt.Errorf("retry-group test failure")
}
func (s *RetryGroupSignal) AutoRetry() bool           { return true }
func (s *RetryGroupSignal) RetryGroup() bool          { return true }
func (s *RetryGroupSignal) MaxRetries() int           { return 2 }
func (s *RetryGroupSignal) SleepAfter() time.Duration { return time.Second }

var _ signal.SignalWithAutoRetry = (*RetryGroupSignal)(nil)
var _ signal.SignalWithRetryGroup = (*RetryGroupSignal)(nil)
var _ signal.SignalWithMaxRetries = (*RetryGroupSignal)(nil)

// --- CountdownSignal: fails N times then succeeds ---
// Uses SignalWithStepContext + GetFlowStep activity to read the step's
// RetryIndex from DB, then succeeds when RetryIndex >= SucceedAtRetry.

const CountdownSignalType signal.SignalType = "test-flow-countdown"

type CountdownSignal struct {
	SucceedAtRetry int    `json:"succeed_at_retry"`
	StepID         string `json:"step_id,omitempty"`
	FlowID         string `json:"flow_id,omitempty"`
}

func (s *CountdownSignal) Type() signal.SignalType         { return CountdownSignalType }
func (s *CountdownSignal) Validate(workflow.Context) error { return nil }
func (s *CountdownSignal) AutoRetry() bool                 { return true }
func (s *CountdownSignal) MaxRetries() int                 { return s.SucceedAtRetry + 2 }
func (s *CountdownSignal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

func (s *CountdownSignal) Execute(ctx workflow.Context) error {
	if s.StepID == "" {
		return fmt.Errorf("countdown signal: no step context")
	}

	// Look up the step from DB to check its RetryIndex.
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return fmt.Errorf("countdown signal: unable to get step: %w", err)
	}

	if step.RetryIndex >= s.SucceedAtRetry {
		return nil // success
	}
	return fmt.Errorf("countdown signal: retry %d < target %d", step.RetryIndex, s.SucceedAtRetry)
}

func (s *CountdownSignal) SleepAfter() time.Duration { return time.Second }

var _ signal.SignalWithAutoRetry = (*CountdownSignal)(nil)
var _ signal.SignalWithMaxRetries = (*CountdownSignal)(nil)
var _ signal.SignalWithStepContext = (*CountdownSignal)(nil)

// --- CountdownGroupSignal: like CountdownSignal but requests group retry ---

const CountdownGroupSignalType signal.SignalType = "test-flow-countdown-group"

type CountdownGroupSignal struct {
	SucceedAtRetry int    `json:"succeed_at_retry"`
	StepID         string `json:"step_id,omitempty"`
	FlowID         string `json:"flow_id,omitempty"`
}

func (s *CountdownGroupSignal) Type() signal.SignalType         { return CountdownGroupSignalType }
func (s *CountdownGroupSignal) Validate(workflow.Context) error { return nil }
func (s *CountdownGroupSignal) AutoRetry() bool                 { return true }
func (s *CountdownGroupSignal) RetryGroup() bool                { return true }
func (s *CountdownGroupSignal) MaxRetries() int                 { return s.SucceedAtRetry + 2 }
func (s *CountdownGroupSignal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

func (s *CountdownGroupSignal) Execute(ctx workflow.Context) error {
	if s.StepID == "" {
		return fmt.Errorf("countdown-group signal: no step context")
	}

	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return fmt.Errorf("countdown-group signal: unable to get step: %w", err)
	}

	if step.RetryIndex >= s.SucceedAtRetry {
		return nil
	}
	return fmt.Errorf("countdown-group signal: retry %d < target %d", step.RetryIndex, s.SucceedAtRetry)
}

func (s *CountdownGroupSignal) SleepAfter() time.Duration { return time.Second }

var _ signal.SignalWithAutoRetry = (*CountdownGroupSignal)(nil)
var _ signal.SignalWithRetryGroup = (*CountdownGroupSignal)(nil)
var _ signal.SignalWithMaxRetries = (*CountdownGroupSignal)(nil)
var _ signal.SignalWithStepContext = (*CountdownGroupSignal)(nil)

// --- PlanApplyFailSignal: "apply" signal that always fails and requests group retry.
// Implements CloneSteps to return both a plan (SuccessSignal) and apply (self) step
// when retried, simulating a plan+apply retry pattern.

const PlanApplyFailSignalType signal.SignalType = "test-flow-plan-apply-fail"

type PlanApplyFailSignal struct{}

func init() {
	catalog.Register(PlanApplyFailSignalType, func() signal.Signal { return &PlanApplyFailSignal{} })
}

func (s *PlanApplyFailSignal) Type() signal.SignalType         { return PlanApplyFailSignalType }
func (s *PlanApplyFailSignal) Validate(workflow.Context) error { return nil }
func (s *PlanApplyFailSignal) Execute(workflow.Context) error {
	return fmt.Errorf("plan-apply-fail: apply failed")
}
func (s *PlanApplyFailSignal) AutoRetry() bool           { return true }
func (s *PlanApplyFailSignal) RetryGroup() bool          { return true }
func (s *PlanApplyFailSignal) MaxRetries() int           { return 2 }
func (s *PlanApplyFailSignal) SleepAfter() time.Duration { return time.Second }

func (s *PlanApplyFailSignal) Clone(_ workflow.Context, originalStepName string) ([]signal.CloneStepDef, error) {
	return []signal.CloneStepDef{
		{
			Name:          originalStepName + "-plan",
			Signal:        &SuccessSignal{},
			ExecutionType: "system",
		},
		{
			Name:          originalStepName + "-apply",
			Signal:        &PlanApplyFailSignal{},
			ExecutionType: "system",
		},
	}, nil
}

var _ signal.SignalWithAutoRetry = (*PlanApplyFailSignal)(nil)
var _ signal.SignalWithRetryGroup = (*PlanApplyFailSignal)(nil)
var _ signal.SignalWithMaxRetries = (*PlanApplyFailSignal)(nil)
var _ signal.SignalWithCloneSteps = (*PlanApplyFailSignal)(nil)

// --- ManualRetryGroupCountdownSignal: auto-retries once (group retry), then
// requires manual retry to succeed. Uses SignalWithRetryCount to branch on
// the group retry generation.
//
// MaxRetries=2, AutoRetry=true, RetryGroup=true.
// - GroupRetryCount < 2: fail (auto-retry produces generation 1, which also fails)
// - GroupRetryCount >= 2: succeed (manual retry via RetryStep creates generation 2)

const ManualRetryGroupCountdownSignalType signal.SignalType = "test-flow-manual-retry-group-countdown"

type ManualRetryGroupCountdownSignal struct {
	StepID          string `json:"step_id,omitempty"`
	FlowID          string `json:"flow_id,omitempty"`
	GroupRetryCount int    `json:"-"`
}

func init() {
	catalog.Register(ManualRetryGroupCountdownSignalType, func() signal.Signal { return &ManualRetryGroupCountdownSignal{} })
}

func (s *ManualRetryGroupCountdownSignal) Type() signal.SignalType {
	return ManualRetryGroupCountdownSignalType
}
func (s *ManualRetryGroupCountdownSignal) Validate(workflow.Context) error { return nil }
func (s *ManualRetryGroupCountdownSignal) AutoRetry() bool                 { return true }
func (s *ManualRetryGroupCountdownSignal) RetryGroup() bool                { return true }
func (s *ManualRetryGroupCountdownSignal) MaxRetries() int                 { return 2 }
func (s *ManualRetryGroupCountdownSignal) SleepAfter() time.Duration       { return time.Second }
func (s *ManualRetryGroupCountdownSignal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}
func (s *ManualRetryGroupCountdownSignal) SetRetryCount(retryIndex, groupRetryIndex int) {
	s.GroupRetryCount = groupRetryIndex
}

func (s *ManualRetryGroupCountdownSignal) Execute(workflow.Context) error {
	if s.GroupRetryCount >= 2 {
		return nil // success on manual retry
	}
	return fmt.Errorf("manual-retry-group-countdown: group retry %d < 2", s.GroupRetryCount)
}

var _ signal.SignalWithAutoRetry = (*ManualRetryGroupCountdownSignal)(nil)
var _ signal.SignalWithRetryGroup = (*ManualRetryGroupCountdownSignal)(nil)
var _ signal.SignalWithMaxRetries = (*ManualRetryGroupCountdownSignal)(nil)
var _ signal.SignalWithStepContext = (*ManualRetryGroupCountdownSignal)(nil)
var _ signal.SignalWithRetryCount = (*ManualRetryGroupCountdownSignal)(nil)

// --- CancellableTestSignal: blocks until cancelled, writes marker on Cancel() ---
// This proves the Cancel() method was called by writing to the step's
// ResultDirective field (a known, queryable column).

const CancellableTestSignalType signal.SignalType = "test-flow-cancellable"

// CancelMarker is the value written to ResultDirective when Cancel() is invoked.
const CancelMarker = "cancel-callback-invoked"

type CancellableTestSignal struct {
	StepID string `json:"step_id,omitempty"`
	FlowID string `json:"flow_id,omitempty"`
}

func init() {
	catalog.Register(CancellableTestSignalType, func() signal.Signal { return &CancellableTestSignal{} })
}

func (s *CancellableTestSignal) Type() signal.SignalType         { return CancellableTestSignalType }
func (s *CancellableTestSignal) Validate(workflow.Context) error { return nil }
func (s *CancellableTestSignal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

func (s *CancellableTestSignal) Execute(ctx workflow.Context) error {
	// Block forever until the context is cancelled
	return workflow.Await(ctx, func() bool {
		return ctx.Err() != nil
	})
}

func (s *CancellableTestSignal) Cancel(ctx workflow.Context) error {
	if s.StepID == "" {
		return nil
	}
	// Write CancelMarker to the step's ResultDirective so the test can verify
	// that Cancel() was actually invoked at this level.
	return activities.AwaitPkgWorkflowsFlowUpdateFlowStepResultDirective(ctx, activities.UpdateFlowStepResultDirectiveRequest{
		StepID:    s.StepID,
		Directive: CancelMarker,
	})
}

func (s *CancellableTestSignal) SleepAfter() time.Duration { return time.Second }

var _ signal.Signal = (*CancellableTestSignal)(nil)
var _ signal.SignalWithCancel = (*CancellableTestSignal)(nil)
var _ signal.SignalWithStepContext = (*CancellableTestSignal)(nil)
