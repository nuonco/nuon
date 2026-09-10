package generateworkflowsteps

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const SignalType qsignal.SignalType = "generate-workflow-steps"

// cancelFinishedAtVersion gates the finished-at write on cancellation so
// in-flight histories, which never scheduled it, replay deterministically.
// todo(sk): clean this after teminating old workflows
const cancelFinishedAtVersion = "generate-steps-cancel-finished-at-v1"

// earlyEagerPublishVersion gates handing the generator an eager publisher.
// Publishing mid-generation completes the eager-step-groups update earlier in
// history relative to the generator's activity commands, so in-flight
// generate-steps executions must keep consuming the eager subset only once
// gen() has returned.
// todo(sk): clean this after terminating old workflows
const earlyEagerPublishVersion = "generate-steps-early-eager-publish-v1"

// generatorRegistry is populated at init time by packages that register
// step generators for specific owner types. This avoids import cycles.
var generatorRegistry = map[string]func() map[app.WorkflowType]flow.WorkflowStepGenerator{}

// RegisterGenerators registers a step generator factory for an owner type.
// Called from init() functions to avoid import cycles.
func RegisterGenerators(ownerType string, factory func() map[app.WorkflowType]flow.WorkflowStepGenerator) {
	generatorRegistry[ownerType] = factory
}

type Signal struct {
	WorkflowID string `json:"workflow_id"`
	OwnerType  string `json:"owner_type"`

	// result is populated by Execute and read by the FetchSteps handler.
	result *app.GenerateStepsResult
	done   bool
	err    error

	// eagerStepGroups holds step groups that can be executed immediately while
	// remaining groups are still generating. Set before done to allow early consumption.
	eagerStepGroups      *app.GenerateStepsResult
	eagerStepGroupsReady bool

	// fetched tracks whether each update handler has been called, so Execute
	// can block until all consumers have retrieved results before completing.
	fetchStepsCalled      bool
	eagerStepGroupsCalled bool
}

var (
	_ qsignal.Signal                     = (*Signal)(nil)
	_ qsignal.SignalWithUpdateHandlers   = (*Signal)(nil)
	_ qsignal.SignalWithFetchSteps       = (*Signal)(nil)
	_ qsignal.SignalWithLifecycleContext = (*Signal)(nil)
)

func (s *Signal) SetWorkflowID(id string) {
	s.WorkflowID = id
}

func (s *Signal) LifecycleContext() qsignal.SignalLifecycleContext {
	return qsignal.SignalLifecycleContext{
		Operation:  "generate-workflow-steps",
		WorkflowID: s.WorkflowID,
		OwnerType:  s.OwnerType,
	}
}

func (s *Signal) Type() qsignal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Look up the workflow and generate steps.
	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		s.err = errors.Wrap(err, "unable to get workflow")
		s.done = true
		return s.err
	}

	defer func() {
		if ctx.Err() == nil {
			return
		}
		// new disconnected context because current context has already been cancelled due to workflow cancellation
		dCtx, dCancel := workflow.NewDisconnectedContext(ctx)
		defer dCancel()
		// In-flight histories never scheduled this activity on cancel.
		if workflow.GetVersion(dCtx, cancelFinishedAtVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
			return
		}
		_ = workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(dCtx, flw.ID)
	}()

	// Resolve owner type — prefer the signal field, fall back to the workflow.
	ownerType := s.OwnerType
	if ownerType == "" {
		ownerType = flw.OwnerType
	}

	factory, has := generatorRegistry[ownerType]
	if !has {
		s.err = errors.Errorf("no generators registered for owner type %s", ownerType)
		s.done = true
		return s.err
	}

	gens := factory()
	gen, has := gens[flw.Type]
	if !has {
		s.err = errors.Errorf("no step generator for workflow type %s", flw.Type)
		s.done = true
		return s.err
	}

	// Hand the generator a publisher so it can release its eager groups at its
	// own boundary instead of forcing the conductor to wait for the whole
	// generation. Without this the eager-step-groups handler unblocks only
	// after gen() returns, so nothing starts early and the optimization is
	// inert.
	genCtx := ctx
	if workflow.GetVersion(ctx, earlyEagerPublishVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
		genCtx = flow.WithEagerPublisher(ctx, func(groups []*app.WorkflowStepGroup, steps []*app.WorkflowStep) {
			if s.eagerStepGroupsReady {
				return
			}
			s.eagerStepGroups = &app.GenerateStepsResult{Steps: steps, Groups: groups}
			s.eagerStepGroupsReady = true
		})
	}

	result, err := gen(genCtx, flw)
	if err != nil {
		s.err = errors.Wrapf(err, "unable to generate steps for workflow %s", flw.ID)
		s.done = true
		return s.err
	}

	// Fallback for a generator that published nothing: consume the eager subset
	// of the finished result, which is what every caller did before generators
	// could publish early.
	if !s.eagerStepGroupsReady && len(result.Groups) > 0 {
		eagerGroups, eagerSteps := flow.EagerSubset(result.Groups, result.Steps)
		s.eagerStepGroups = &app.GenerateStepsResult{
			Steps:  eagerSteps,
			Groups: eagerGroups,
		}
	}
	s.eagerStepGroupsReady = true

	s.result = result
	s.done = true

	// Wait for both update handlers to be called so the conductor can
	// retrieve the generated steps before this signal completes.
	return workflow.Await(ctx, func() bool {
		return s.fetchStepsCalled && s.eagerStepGroupsCalled
	})
}

func (s *Signal) RegisterUpdateHandlers(ctx workflow.Context) error {
	if err := workflow.SetUpdateHandlerWithOptions(
		ctx, "eager-step-groups",
		func(ctx workflow.Context) (*app.GenerateStepsResult, error) {
			defer func() { s.eagerStepGroupsCalled = true }()
			// Block until eager step groups are ready.
			if err := workflow.Await(ctx, func() bool { return s.eagerStepGroupsReady }); err != nil {
				return nil, err
			}
			if s.err != nil {
				return nil, s.err
			}
			return s.eagerStepGroups, nil
		},
		workflow.UpdateHandlerOptions{},
	); err != nil {
		return err
	}

	return workflow.SetUpdateHandlerWithOptions(
		ctx, "FetchSteps",
		func(ctx workflow.Context) (*app.GenerateStepsResult, error) {
			defer func() { s.fetchStepsCalled = true }()
			// Block until Execute has finished generating steps.
			if err := workflow.Await(ctx, func() bool { return s.done }); err != nil {
				return nil, err
			}
			if s.err != nil {
				return nil, s.err
			}
			return s.result, nil
		},
		workflow.UpdateHandlerOptions{},
	)
}
