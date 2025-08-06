package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
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

type stepGroup struct {
	idx int
}

func newStepGroup() *stepGroup {
	return &stepGroup{}
}

func (s *stepGroup) nextGroup() {
	s.idx++
}

func (s *stepGroup) installSignalStep(ctx workflow.Context, installID, name string, metadata pgtype.Hstore, signal *signals.Signal, planOnly bool, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	opts = append(opts, WithGroupIdx(s.idx))
	return installSignalStep(ctx, installID, name, metadata, signal, planOnly, opts...)
}
