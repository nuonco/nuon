package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type stepGroup struct {
	idx          int
	groups       []*app.WorkflowStepGroup
	currentGroup *app.WorkflowStepGroup
}

func newStepGroup() *stepGroup {
	return &stepGroup{}
}

func (s *stepGroup) nextGroup() {
	s.nextGroupWithOpts("", false)
}

func (s *stepGroup) nextGroupParallel() {
	s.nextGroupWithOpts("", true)
}

func (s *stepGroup) nextGroupNamed(name string) {
	s.nextGroupWithOpts(name, false)
}

func (s *stepGroup) nextGroupWithOpts(name string, parallel bool) {
	s.nextGroupFull(name, parallel, nil)
}

func (s *stepGroup) nextGroupLabeled(lbls labels.Labels) {
	s.nextGroupFull("", false, lbls)
}

func (s *stepGroup) nextGroupFull(name string, parallel bool, lbls labels.Labels) {
	s.idx++
	g := &app.WorkflowStepGroup{
		GroupIdx: s.idx,
		Parallel: parallel,
		Name:     name,
		Status:   app.CompositeStatus{Status: app.StatusPending},
	}
	if lbls != nil {
		g.Labels = lbls
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

func (s *stepGroup) installSignalStep(ctx workflow.Context, installID, name string, metadata pgtype.Hstore, sig signal.Signal, planOnly bool, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	opts = append(opts, WithGroupIdx(s.idx))
	return installSignalStep(ctx, installID, name, metadata, sig, planOnly, opts...)
}
