package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type AppendStepRequest struct {
	Name           string                        `json:"name"`
	Signal         signaldb.SignalData           `json:"signal"`
	ExecutionType  app.WorkflowStepExecutionType `json:"execution_type,omitempty"`
	StepTargetType string                        `json:"step_target_type,omitempty"`
	StepTargetID   string                        `json:"step_target_id,omitempty"`
	Retryable      bool                          `json:"retryable,omitempty"`
	Skippable      bool                          `json:"skippable,omitempty"`
	SkipOnFailure  bool                          `json:"skip_on_failure,omitempty"`
}

type AppendStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	GroupID    string `json:"group_id"`
	StepID     string `json:"step_id"`
}

func (s *Signal) appendStepHandler(ctx workflow.Context, req AppendStepRequest) (*AppendStepResponse, error) {
	defer s.beginUpdate()()

	if !s.Resident {
		return nil, fmt.Errorf("append-step is only valid on a resident workflow")
	}
	if req.Signal.Signal == nil {
		return nil, fmt.Errorf("append-step requires a signal")
	}

	groups, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroups(ctx, s.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to load step groups: %w", err)
	}
	nextGroupIdx := len(groups)

	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to load steps: %w", err)
	}
	nextStepIdx := len(steps)

	if req.StepTargetID != "" {
		for i := range steps {
			if steps[i].StepTargetID == req.StepTargetID {
				s.wakeForAppend()
				return &AppendStepResponse{
					WorkflowID: s.WorkflowID,
					GroupID:    steps[i].WorkflowStepGroupID,
					StepID:     steps[i].ID,
				}, nil
			}
		}
	}

	createdGroups, err := workflowactivities.AwaitPkgWorkflowsFlowCreateFlowStepGroups(ctx, workflowactivities.CreateFlowStepGroupsRequest{
		Groups: []workflowactivities.CreateFlowStepGroup{{
			WorkflowID: s.WorkflowID,
			GroupIdx:   nextGroupIdx,
			Name:       req.Name,
			Status:     app.NewCompositeTemporalStatus(ctx, app.StatusPending),
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create appended group: %w", err)
	}
	group := createdGroups[0]

	execType := req.ExecutionType
	if execType == "" {
		execType = app.WorkflowStepExecutionTypeSystem
	}
	sig := req.Signal

	createdSteps, err := workflowactivities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, workflowactivities.CreateFlowStepsRequest{
		Steps: []workflowactivities.CreateFlowStep{{
			FlowID:              s.WorkflowID,
			OwnerID:             s.OwnerID,
			OwnerType:           s.OwnerType,
			Status:              app.NewCompositeTemporalStatus(ctx, app.StatusPending),
			Name:                req.Name,
			Idx:                 nextStepIdx,
			GroupIdx:            nextGroupIdx,
			WorkflowStepGroupID: group.ID,
			ExecutionType:       execType,
			StepTargetType:      req.StepTargetType,
			StepTargetID:        req.StepTargetID,
			Retryable:           req.Retryable,
			Skippable:           req.Skippable,
			SkipOnFailure:       req.SkipOnFailure,
			QueueSignal:         &sig,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create appended step: %w", err)
	}
	step := createdSteps[0]

	s.wakeForAppend()

	return &AppendStepResponse{
		WorkflowID: s.WorkflowID,
		GroupID:    group.ID,
		StepID:     step.ID,
	}, nil
}

func (s *Signal) wakeForAppend() {
	s.appendRequested = true
	s.resumeRunType = app.WorkflowRunTypeResume
}
