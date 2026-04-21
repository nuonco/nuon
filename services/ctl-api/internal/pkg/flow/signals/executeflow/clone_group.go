package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// cloneGroupForRetry clones all steps in the given group. It marks existing
// steps as Discarded and creates new clones with an incremented GroupRetryIdx.
// When step groups exist, a new WorkflowStepGroup is created and the old one
// is marked as discarded.
func (s *Signal) cloneGroupForRetry(ctx workflow.Context, groupIdx int) error {
	allSteps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return fmt.Errorf("unable to get flow steps: %w", err)
	}

	var groupSteps []app.WorkflowStep
	for _, step := range allSteps {
		if step.GroupIdx == groupIdx {
			groupSteps = append(groupSteps, step)
		}
	}
	if len(groupSteps) == 0 {
		return fmt.Errorf("no steps found in group %d", groupIdx)
	}

	maxIdx := 0
	newGroupRetryIdx := 0
	for _, step := range groupSteps {
		if step.Idx > maxIdx {
			maxIdx = step.Idx
		}
		if step.GroupRetryIdx >= newGroupRetryIdx {
			newGroupRetryIdx = step.GroupRetryIdx + 1
		}
	}

	// If steps have a WorkflowStepGroupID, create a new group and mark the old one as discarded.
	var newGroupID string
	oldGroupID := groupSteps[0].WorkflowStepGroupID
	if oldGroupID != "" {
		// Mark old group as discarded.
		_ = statusactivities.AwaitPkgStatusUpdateFlowStepGroupStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: oldGroupID,
			Status: app.CompositeStatus{
				Status: app.StatusDiscarded,
				Metadata: map[string]any{
					"reason": "group retry",
				},
			},
		})

		// Create a new group for the retry.
		newGroupID = domains.NewWorkflowStepGroupID()
		if _, err := workflowactivities.AwaitPkgWorkflowsFlowCreateFlowStepGroups(ctx, workflowactivities.CreateFlowStepGroupsRequest{
			Groups: []workflowactivities.CreateFlowStepGroup{
				{
					ID:         newGroupID,
					WorkflowID: s.WorkflowID,
					GroupIdx:   groupIdx,
					Parallel:   groupSteps[0].GroupParallel,
					Status:     app.NewCompositeTemporalStatus(ctx, app.StatusPending),
				},
			},
		}); err != nil {
			return fmt.Errorf("unable to create retry group: %w", err)
		}
	}

	// Mark all existing steps as discarded
	for _, step := range groupSteps {
		if step.Status.Status == app.StatusDiscarded {
			continue
		}
		_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusDiscarded,
				Metadata: map[string]any{
					"reason": "group retry",
				},
			},
		})
	}

	// Only clone primary steps (no retries) to avoid duplicates.
	var primarySteps []app.WorkflowStep
	for _, step := range groupSteps {
		if step.RetryIndex > 0 || step.GroupRetryIdx > 0 {
			continue
		}
		primarySteps = append(primarySteps, step)
	}

	// Clone each primary step
	cloneSteps := make([]workflowactivities.CreateFlowStep, 0, len(primarySteps))
	for i, step := range primarySteps {
		var qs *signaldb.SignalData
		if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
			qs = &signaldb.SignalData{Signal: step.QueueSignal.Signal}
		}

		cloneSteps = append(cloneSteps, workflowactivities.CreateFlowStep{
			FlowID:              s.WorkflowID,
			OwnerID:             step.OwnerID,
			OwnerType:           step.OwnerType,
			Name:                step.Name,
			Signal:              step.Signal,
			QueueSignal:         qs,
			Status:              app.NewCompositeTemporalStatus(ctx, app.StatusPending),
			Idx:                 maxIdx + 100 + i,
			ExecutionType:       step.ExecutionType,
			Metadata:            step.Metadata,
			Retryable:           step.Retryable,
			Skippable:           step.Skippable,
			GroupIdx:            step.GroupIdx,
			GroupRetryIdx:       newGroupRetryIdx,
			StepTargetType:      step.StepTargetType,
			RetryIndex:          0,
			WorkflowStepGroupID: newGroupID,
		})
	}

	if _, err := workflowactivities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, workflowactivities.CreateFlowStepsRequest{
		Steps: cloneSteps,
	}); err != nil {
		return fmt.Errorf("unable to create clone steps: %w", err)
	}

	return nil
}
