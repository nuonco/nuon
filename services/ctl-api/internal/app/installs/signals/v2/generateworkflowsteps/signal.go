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

	// steps is populated by Execute and read by the FetchSteps handler.
	steps []*app.WorkflowStep
	done  bool
	err   error
}

var (
	_ qsignal.Signal                   = (*Signal)(nil)
	_ qsignal.SignalWithUpdateHandlers = (*Signal)(nil)
	_ qsignal.SignalWithFetchSteps     = (*Signal)(nil)
)

func (s *Signal) SetWorkflowID(id string) {
	s.WorkflowID = id
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

	steps, err := gen(ctx, flw)
	if err != nil {
		s.err = errors.Wrapf(err, "unable to generate steps for workflow %s", flw.ID)
		s.done = true
		return s.err
	}

	s.steps = steps
	s.done = true

	// Wait for the FetchSteps update to be called so the conductor can
	// retrieve the generated steps before this signal completes.
	return workflow.Await(ctx, func() bool { return false })
}

func (s *Signal) RegisterUpdateHandlers(ctx workflow.Context) error {
	return workflow.SetUpdateHandlerWithOptions(ctx, "FetchSteps",
		func(ctx workflow.Context) ([]*app.WorkflowStep, error) {
			// Block until Execute has finished generating steps.
			if err := workflow.Await(ctx, func() bool { return s.done }); err != nil {
				return nil, err
			}
			if s.err != nil {
				return nil, s.err
			}
			return s.steps, nil
		},
		workflow.UpdateHandlerOptions{},
	)
}
