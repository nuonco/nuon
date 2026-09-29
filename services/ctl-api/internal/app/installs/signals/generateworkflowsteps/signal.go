package generateworkflowsteps

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const SignalType qsignal.SignalType = "generate-workflow-steps"

// why: cancelFinishedAtVersion gates the finished-at write on cancellation so
// in-flight histories, which never scheduled it, replay deterministically.
// todo(sk): clean this after teminating old workflows
const cancelFinishedAtVersion = "generate-steps-cancel-finished-at-v1"

var generatorRegistry = map[string]func() map[app.WorkflowType]flow.WorkflowStepGenerator{}

func RegisterGenerators(ownerType string, factory func() map[app.WorkflowType]flow.WorkflowStepGenerator) {
	generatorRegistry[ownerType] = factory
}

type Signal struct {
	WorkflowID string `json:"workflow_id"`
	OwnerType  string `json:"owner_type"`

	OrgID     string `json:"org_id,omitempty"`
	OrgName   string `json:"org_name,omitempty"`
	OwnerID   string `json:"owner_id,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`

	result *app.GenerateStepsResult
	done   bool
	err    error

	eagerStepGroups      *app.GenerateStepsResult
	eagerStepGroupsReady bool

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

func (s *Signal) SetLifecycleIdentity(orgID, orgName, ownerID, ownerName string) {
	s.OrgID = orgID
	s.OrgName = orgName
	s.OwnerID = ownerID
	s.OwnerName = ownerName
}

func (s *Signal) LifecycleContext() qsignal.SignalLifecycleContext {
	lc := qsignal.SignalLifecycleContext{
		Operation:  "generate-workflow-steps",
		WorkflowID: s.WorkflowID,
		OrgID:      s.OrgID,
		OrgName:    s.OrgName,
		OwnerID:    s.OwnerID,
		OwnerType:  s.OwnerType,
		OwnerName:  s.OwnerName,
	}
	if s.OwnerType == plugins.TableNameOf[app.Install]() && s.OwnerID != "" {
		lc.InstallID = &s.OwnerID
	}
	return lc
}

func (s *Signal) Type() qsignal.SignalType {
	return SignalType
}

func (s *Signal) SleepAfter() time.Duration { return time.Second }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		s.err = errors.Wrap(err, "unable to get workflow")
		s.done = true
		return s.err
	}
	telemetry := cctx.WorkflowTelemetry{
		OrgID:        flw.OrgID,
		OrgName:      flw.Org.Name,
		WorkflowID:   flw.ID,
		WorkflowType: string(flw.Type),
		OwnerID:      s.OwnerID,
		OwnerType:    flw.OwnerType,
		OwnerName:    flw.OwnerName,
	}
	if flw.OwnerType == plugins.TableNameOf[app.Install]() {
		telemetry.InstallID = s.OwnerID
		telemetry.InstallName = flw.OwnerName
	}
	ctx = cctx.SetWorkflowTelemetryWorkflowContext(ctx, telemetry)

	defer func() {
		if ctx.Err() == nil {
			return
		}
		// why: new disconnected context because current context has already been cancelled due to workflow cancellation
		dCtx, dCancel := workflow.NewDisconnectedContext(ctx)
		defer dCancel()
		// why: In-flight histories never scheduled this activity on cancel.
		if workflow.GetVersion(dCtx, cancelFinishedAtVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
			return
		}
		_ = workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(dCtx, flw.ID)
	}()

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

	result, err := gen(ctx, flw)
	if err != nil {
		s.err = errors.Wrapf(err, "unable to generate steps for workflow %s", flw.ID)
		s.done = true
		return s.err
	}

	if len(result.Groups) > 0 {
		var eagerGroups []*app.WorkflowStepGroup
		for _, g := range result.Groups {
			if g.EagerExecution {
				eagerGroups = append(eagerGroups, g)
			}
		}
		if len(eagerGroups) == 0 {
			eagerGroups = []*app.WorkflowStepGroup{result.Groups[0]}
		}

		eagerGroupIdxs := make(map[int]bool)
		for _, g := range eagerGroups {
			eagerGroupIdxs[g.GroupIdx] = true
		}

		var eagerSteps []*app.WorkflowStep
		for _, step := range result.Steps {
			if eagerGroupIdxs[step.GroupIdx] {
				eagerSteps = append(eagerSteps, step)
			}
		}
		s.eagerStepGroups = &app.GenerateStepsResult{
			Steps:  eagerSteps,
			Groups: eagerGroups,
		}
	}
	s.eagerStepGroupsReady = true

	s.result = result
	s.done = true

	return workflow.Await(ctx, func() bool {
		return s.fetchStepsCalled && s.eagerStepGroupsCalled
	})
}

func (s *Signal) RegisterUpdateHandlers(ctx workflow.Context) error {
	if err := workflow.SetUpdateHandlerWithOptions(
		ctx, "eager-step-groups",
		func(ctx workflow.Context) (*app.GenerateStepsResult, error) {
			defer func() { s.eagerStepGroupsCalled = true }()
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
