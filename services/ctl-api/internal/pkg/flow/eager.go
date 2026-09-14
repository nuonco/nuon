package flow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type eagerPublisherKey struct{}

// EagerPublisher receives the eager step groups as soon as a generator has
// finished building them, rather than when the generator returns.
type EagerPublisher func(groups []*app.WorkflowStepGroup, steps []*app.WorkflowStep)

// WithEagerPublisher installs fn for the generator invoked with the returned
// context.
func WithEagerPublisher(ctx workflow.Context, fn EagerPublisher) workflow.Context {
	return workflow.WithValue(ctx, eagerPublisherKey{}, fn)
}

// PublishEagerGroups hands the eager subset of groups/steps to the installed
// publisher so the conductor can begin executing those groups while the
// generator keeps running. A generator invoked without a publisher — tests, or
// a caller that does not want early execution — is unaffected, and its groups
// are consumed when it returns as before.
//
// Call this at the point where a generator stops producing eager groups and
// starts doing the expensive work for the later ones. That point is earlier
// than the first non-eager group: the generator has to load the app config and
// component graph before it can open that group, and those loads are the bulk
// of what this is trying to move off the critical path.
func PublishEagerGroups(ctx workflow.Context, groups []*app.WorkflowStepGroup, steps []*app.WorkflowStep) {
	fn, ok := ctx.Value(eagerPublisherKey{}).(EagerPublisher)
	if !ok || fn == nil {
		return
	}

	eagerGroups, eagerSteps := EagerSubset(groups, steps)
	if len(eagerGroups) == 0 {
		return
	}

	fn(eagerGroups, eagerSteps)
}

// EagerSubset returns the groups marked EagerExecution and the steps belonging
// to them. When no group carries the flag the first group is treated as eager,
// matching how generated results were consumed before groups had it.
func EagerSubset(groups []*app.WorkflowStepGroup, steps []*app.WorkflowStep) ([]*app.WorkflowStepGroup, []*app.WorkflowStep) {
	if len(groups) == 0 {
		return nil, nil
	}

	var eagerGroups []*app.WorkflowStepGroup
	for _, g := range groups {
		if g.EagerExecution {
			eagerGroups = append(eagerGroups, g)
		}
	}
	if len(eagerGroups) == 0 {
		eagerGroups = []*app.WorkflowStepGroup{groups[0]}
	}

	eagerIdxs := make(map[int]bool, len(eagerGroups))
	for _, g := range eagerGroups {
		eagerIdxs[g.GroupIdx] = true
	}

	var eagerSteps []*app.WorkflowStep
	for _, step := range steps {
		if eagerIdxs[step.GroupIdx] {
			eagerSteps = append(eagerSteps, step)
		}
	}

	return eagerGroups, eagerSteps
}
