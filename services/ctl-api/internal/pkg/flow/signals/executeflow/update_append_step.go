package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// AppendStepRequest is the input for the "append-step" update handler. It adds a
// new single-step group to a parked Resident workflow (e.g. a notebook cell) and
// wakes the execute loop to run only that step.
//
// The caller builds the step metadata (execution type, target type) the same way
// installSignalStep does — this keeps the low-level flow package free of the
// installs signal taxonomy. Queue IDs are intentionally left empty: dispatch
// resolves the target queue from the signal, exactly like the adhoc-action path.
type AppendStepRequest struct {
	Name           string                        `json:"name"`
	Signal         signaldb.SignalData           `json:"signal"`
	ExecutionType  app.WorkflowStepExecutionType `json:"execution_type,omitempty"`
	StepTargetType string                        `json:"step_target_type,omitempty"`
	Retryable      bool                          `json:"retryable,omitempty"`
	Skippable      bool                          `json:"skippable,omitempty"`
}

// AppendStepResponse reports the IDs of the freshly-created group and step so the
// caller (e.g. a notebook cell run) can track the appended work.
type AppendStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	GroupID    string `json:"group_id"`
	StepID     string `json:"step_id"`
}

// appendStepHandler adds a new group + signal-backed step to the end of a parked
// resident workflow and wakes the execute loop to run only that group.
//
// It re-reads groups from the DB to compute the next group index/position, then
// persists a new group and step via the same create activities the conductor
// uses. handle() re-reads groups on each invocation, so resuming from the new
// group position runs only the appended step; isWorkflowComplete() then sees all
// steps terminal again and the loop re-parks.
func (s *Signal) appendStepHandler(ctx workflow.Context, req AppendStepRequest) (*AppendStepResponse, error) {
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
	// Groups are ordered by GroupIdx and contiguous, so the next group's index
	// equals its position in the slice — which is what handle() resumes from.
	nextGroupIdx := len(groups)

	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to load steps: %w", err)
	}
	nextStepIdx := len(steps)

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
			Retryable:           req.Retryable,
			Skippable:           req.Skippable,
			QueueSignal:         &sig,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create appended step: %w", err)
	}
	step := createdSteps[0]

	// Wake the parked resident loop to run only the new group. If the loop is
	// not parked yet (still running an earlier append), the new group is already
	// persisted and will be picked up on the next resume.
	if s.awaitingResume {
		s.appendRequested = true
		s.resumeRunType = app.WorkflowRunTypeResume
		s.resumeStepID = step.ID
		s.resumeStartIdx = nextGroupIdx
	}

	return &AppendStepResponse{
		WorkflowID: s.WorkflowID,
		GroupID:    group.ID,
		StepID:     step.ID,
	}, nil
}
