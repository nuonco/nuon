package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type WorkflowStepOptions func(*app.WorkflowStep)

func WithSkippable(skippable bool) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.Skippable = skippable
	}
}

func WithSkipOnFailure(skipOnFailure bool) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.SkipOnFailure = skipOnFailure
	}
}

func WithGroupIdx(n int) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.GroupIdx = n
	}
}

func WithStepQueueID(queueID string) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.StepQueueID = queueID
	}
}

func WithTargetQueueID(queueID string) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.TargetQueueID = queueID
	}
}

func WithExecutionType(t app.WorkflowStepExecutionType) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.ExecutionType = t
	}
}

type stepGroup struct {
	idx          int
	groups       []*app.WorkflowStepGroup
	currentGroup *app.WorkflowStepGroup
}

func newStepGroup() *stepGroup {
	return &stepGroup{}
}

func (s *stepGroup) nextGroup() {
	s.idx++
	g := &app.WorkflowStepGroup{
		GroupIdx: s.idx,
		Status:   app.CompositeStatus{Status: app.StatusPending},
	}
	s.groups = append(s.groups, g)
	s.currentGroup = g
}

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

func (s *stepGroup) signalStep(ctx workflow.Context, ownerID, ownerType, name string, metadata pgtype.Hstore, sig signal.Signal, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	opts = append(opts, WithGroupIdx(s.idx))
	return newSignalStep(ctx, ownerID, ownerType, name, metadata, sig, opts...)
}

func (s *stepGroup) appBranchSignalStep(ctx workflow.Context, appBranchID, name string, metadata pgtype.Hstore, sig signal.Signal, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	return s.signalStep(ctx, appBranchID, "app_branches", name, metadata, sig, opts...)
}

func (s *stepGroup) appSignalStep(ctx workflow.Context, appID, name string, metadata pgtype.Hstore, sig signal.Signal, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	return s.signalStep(ctx, appID, "apps", name, metadata, sig, opts...)
}

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
		Skippable: signal.IsSkippable(sig),
	}

	for _, o := range opts {
		o(step)
	}

	return step, nil
}
