package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// runnerHealthyStepName gates a phase on the install's runner reporting
// healthy. Named so generators can tell whether the previous step already
// waited on it instead of matching the string in several places.
const runnerHealthyStepName = "runner healthy"

type stepGroup struct {
	idx          int
	groups       []*app.WorkflowStepGroup
	currentGroup *app.WorkflowStepGroup

	// lastStepName is the name of the most recently emitted step, letting a
	// generator skip a gate the preceding phase has already satisfied.
	lastStepName string

	// flw is the install workflow currently being generated. It is used to
	// stamp install workflow identity onto signals that implement
	// signal.SignalWithMutableLifecycleContext, removing the need for a
	// runtime DB lookup in the queue handler.
	flw *app.Workflow
}

func newStepGroup(flw *app.Workflow) *stepGroup {
	return &stepGroup{flw: flw}
}

func (s *stepGroup) nextGroup() {
	s.nextGroupWithOpts("", false)
}

func (s *stepGroup) nextGroupEager() {
	s.nextGroupWithOpts("", false)
	s.currentGroup.EagerExecution = true
}

func (s *stepGroup) nextGroupParallel() {
	s.nextGroupWithOpts("", true)
}

func (s *stepGroup) nextGroupNamed(name string) {
	s.nextGroupWithOpts(name, false)
}

func (s *stepGroup) nextGroupWithOpts(name string, parallel bool) {
	s.idx++
	g := &app.WorkflowStepGroup{
		GroupIdx: s.idx,
		Parallel: parallel,
		Name:     name,
		Status:   app.CompositeStatus{Status: app.StatusPending},
	}
	s.groups = append(s.groups, g)
	s.currentGroup = g
}

// needsRunnerHealthyGate reports whether a phase about to run on the install's
// runner still has to wait for it. False when the preceding step was already
// that wait, where a second one can only re-confirm the same result.
func (s *stepGroup) needsRunnerHealthyGate() bool {
	return s.lastStepName != runnerHealthyStepName
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

	// Stamp install workflow identity onto the signal so the queue handler can
	// emit lifecycle events without a separate DB lookup.
	if s.flw != nil && sig != nil {
		if mlc, ok := sig.(signal.SignalWithMutableLifecycleContext); ok {
			mlc.SetLifecycleWorkflow(s.flw.ID, string(s.flw.Type))
		}
	}

	step, err := installSignalStep(ctx, installID, name, metadata, sig, planOnly, opts...)
	if err != nil {
		return nil, err
	}
	s.lastStepName = name

	return step, nil
}
