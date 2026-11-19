package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// runnerHealthyStepName gates a phase on the install's runner reporting healthy.
const runnerHealthyStepName = app.RunnerHealthyStepName

// runnerHealthyGateVersion guards the decision to omit the runner-healthy step
// entirely. In-flight workflow histories that already recorded this step
// unconditionally must keep recording it on replay.
const runnerHealthyGateVersion = "runner-healthy-gate-v1"

// shouldGenerateRunnerHealthyStep decides whether to emit the runner-healthy
// gate for an install.
//
//   - appliesStack=true is for workflows that apply the stack (provision,
//     reprovision_stack, stack-changing branch config updates): a stack apply
//     can flip whether a runner exists, so the step is included unless BOTH
//     the desired use_runner AND the actual runner status already agree
//     there's no runner.
//   - appliesStack=false is for every other call site: there's no stack
//     change to wait out, so the step is included unless the actual runner
//     status is not-used.
func shouldGenerateRunnerHealthyStep(ctx workflow.Context, installID string, appliesStack bool) (bool, error) {
	info, err := activities.AwaitGetRunnerHealthyGateInfo(ctx, activities.GetRunnerHealthyGateInfoRequest{
		InstallID: installID,
	})
	if err != nil {
		return false, err
	}

	notUsed := info.RunnerStatus == app.RunnerStatusNotUsed
	if appliesStack {
		return !(notUsed && !info.UsesRunner), nil
	}
	return !notUsed, nil
}

// maybeRunnerHealthyStep generates the runner-healthy gate step for installID
// unless shouldGenerateRunnerHealthyStep decides to omit it, in which case it
// returns a nil step and nil error -- callers should skip appending it. See
// shouldGenerateRunnerHealthyStep for the appliesStack rule.
func maybeRunnerHealthyStep(ctx workflow.Context, sg *stepGroup, installID string, mode awaitrunnerhealthy.Mode, appliesStack, planOnly bool) (*app.WorkflowStep, error) {
	include := true
	if workflow.GetVersion(ctx, runnerHealthyGateVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
		var err error
		include, err = shouldGenerateRunnerHealthyStep(ctx, installID, appliesStack)
		if err != nil {
			return nil, err
		}
	}
	if !include {
		return nil, nil
	}

	return sg.installSignalStep(ctx, installID, runnerHealthyStepName, pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
		Mode:      mode,
	}, planOnly)
}

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
		if mlc, ok := sig.(signal.SignalWithMutableLifecycleContext); ok {
			mlc.SetLifecycleWorkflow(s.flw.ID, string(s.flw.Type))
		}
	}

	return installSignalStep(ctx, installID, name, metadata, sig, planOnly, opts...)
}
