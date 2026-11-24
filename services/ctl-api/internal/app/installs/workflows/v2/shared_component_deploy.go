package v2

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/executeactionworkflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// runnerHealthyStepName gates a phase on the install's runner reporting healthy.
const runnerHealthyStepName = app.RunnerHealthyStepName

type stepGroup struct {
	idx          int
	groups       []*app.WorkflowStepGroup
	currentGroup *app.WorkflowStepGroup

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
		targets, err := s.flw.ExecutionTargets()
		if err != nil {
			return nil, err
		}
		if action, ok := sig.(*executeactionworkflow.Signal); ok && targets != nil && action.AdhocActionRunID == "" {
			found := false
			for _, target := range targets.Actions {
				if target.InstallActionWorkflowID == action.InstallActionWorkflowID {
					action.AppConfigID = targets.AppConfigID
					action.ActionWorkflowConfigID = target.ConfigID
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("action is not in the prepared runbook targets")
			}
		}
		if mlc, ok := sig.(signal.SignalWithMutableLifecycleContext); ok {
			mlc.SetLifecycleWorkflow(s.flw.ID, string(s.flw.Type))
		}
	}

	return installSignalStep(ctx, installID, name, metadata, sig, planOnly, opts...)
}
