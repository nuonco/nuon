package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// WorkflowStepOptions is a functional option for configuring WorkflowStep
type WorkflowStepOptions func(*app.WorkflowStep)

// WithSkippable sets whether a step can be skipped
func WithSkippable(skippable bool) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.Skippable = skippable
	}
}

// WithGroupIdx sets the group index for a step
func WithGroupIdx(n int) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.GroupIdx = n
	}
}

// WithStepQueueID sets the queue where the execute-workflow-step signal runs.
func WithStepQueueID(queueID string) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.StepQueueID = queueID
	}
}

// WithTargetQueueID sets the queue where the inner signal runs.
func WithTargetQueueID(queueID string) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.TargetQueueID = queueID
	}
}

// WithExecutionType overrides the default execution type for a step.
func WithExecutionType(t app.WorkflowStepExecutionType) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.ExecutionType = t
	}
}

// stepGroup tracks the current group index for workflow steps
type stepGroup struct {
	idx          int
	groups       []*app.WorkflowStepGroup
	currentGroup *app.WorkflowStepGroup
}

// newStepGroup creates a new step group starting at index 0
func newStepGroup() *stepGroup {
	return &stepGroup{}
}

// nextGroup increments the group index for sequential execution
func (s *stepGroup) nextGroup() {
	s.idx++
	g := &app.WorkflowStepGroup{
		GroupIdx: s.idx,
		Status:   app.CompositeStatus{Status: app.StatusPending},
	}
	s.groups = append(s.groups, g)
	s.currentGroup = g
}

// nextParallelGroup increments the group index and creates a parallel group.
func (s *stepGroup) nextParallelGroup(name string) {
	s.idx++
	g := &app.WorkflowStepGroup{
		GroupIdx: s.idx,
		Parallel: true,
		Name:     name,
		Status:   app.CompositeStatus{Status: app.StatusPending},
	}
	s.groups = append(s.groups, g)
	s.currentGroup = g
}

func (s *stepGroup) Groups() []*app.WorkflowStepGroup {
	return s.groups
}

func (s *stepGroup) Result(steps []*app.WorkflowStep) *app.GenerateStepsResult {
	return &app.GenerateStepsResult{
		Steps:  steps,
		Groups: s.groups,
	}
}

// signalStep creates a WorkflowStep from a signal with configurable owner type.
func (s *stepGroup) signalStep(ctx workflow.Context, ownerID, ownerType, name string, metadata pgtype.Hstore, sig signal.Signal, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	opts = append(opts, WithGroupIdx(s.idx))
	return newSignalStep(ctx, ownerID, ownerType, name, metadata, sig, opts...)
}

// appBranchSignalStep creates a WorkflowStep from an app branch signal
func (s *stepGroup) appBranchSignalStep(ctx workflow.Context, appBranchID, name string, metadata pgtype.Hstore, sig signal.Signal, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	return s.signalStep(ctx, appBranchID, "app_branches", name, metadata, sig, opts...)
}

// newSignalStep creates a workflow step from a signal with configurable owner type
func newSignalStep(ctx workflow.Context, ownerID, ownerType, name string, metadata pgtype.Hstore, sig signal.Signal, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	if sig == nil {
		step := &app.WorkflowStep{
			Name:          name,
			ExecutionType: app.WorkflowStepExecutionTypeSkipped,
			Status:        app.NewCompositeTemporalStatus(ctx, app.StatusPending),
			Metadata:      metadata,
		}

		for _, o := range opts {
			o(step)
		}

		return step, nil
	}

	step := &app.WorkflowStep{
		Name:           name,
		ExecutionType:  app.WorkflowStepExecutionTypeSystem,
		StepTargetType: ownerType,
		OwnerID:        ownerID,
		OwnerType:      ownerType,
		Status:         app.NewCompositeTemporalStatus(ctx, app.StatusPending),
		Metadata:       metadata,
		QueueSignal: &signaldb.SignalData{
			Signal: sig,
		},
		Retryable: true,
		Skippable: true,
	}

	for _, o := range opts {
		o(step)
	}

	return step, nil
}
