package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"go.temporal.io/sdk/workflow"
)

type WorkflowStepOptions func(*app.WorkflowStep)

func WithSkippable(skippable bool) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.Skippable = skippable
	}
}

func WithGroupIdx(n int) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.GroupIdx = n
	}
}

func WithExecutionType(executionType app.WorkflowStepExecutionType) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.ExecutionType = executionType
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

func (s *stepGroup) Result(steps []*app.WorkflowStep) *app.GenerateStepsResult {
	return &app.GenerateStepsResult{
		Steps:  steps,
		Groups: s.groups,
	}
}

func (s *stepGroup) installSignalStep(ctx workflow.Context, installID, name string, metadata pgtype.Hstore, signal *signals.Signal, planOnly bool, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	opts = append(opts, WithGroupIdx(s.idx))
	return installSignalStep(ctx, installID, name, metadata, signal, planOnly, opts...)
}
