package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstepgroup"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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

	// If steps have a WorkflowStepGroupID, create a new group and mark ALL old ones as discarded.
	// On retry-of-retry, steps from multiple group generations coexist for the same GroupIdx.
	// We must discard every previous group object, not just one.
	var newGroupID string
	groupIDs := make(map[string]bool)
	for _, step := range groupSteps {
		if step.WorkflowStepGroupID != "" {
			groupIDs[step.WorkflowStepGroupID] = true
		}
	}

	if len(groupIDs) > 0 {
		// Mark all old groups as discarded (idempotent — already-discarded is harmless).
		for gid := range groupIDs {
			_ = statusactivities.AwaitPkgStatusUpdateFlowStepGroupStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: gid,
				Status: app.CompositeStatus{
					Status: app.StatusDiscarded,
					Metadata: map[string]any{
						"reason": "group retry",
					},
				},
			})
		}

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

	// Mark all existing steps as retried. Keep the existing status so
	// errors still show correctly in the dashboard.
	for _, step := range groupSteps {
		if step.Status.Status == app.StatusDiscarded {
			continue
		}
		_ = workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowStepRetried(ctx, workflowactivities.UpdateFlowStepRetriedRequest{
			StepID: step.ID,
		})
	}

	// Clone from the latest group retry generation (not always gen 0).
	// On retry-of-retry the latest generation's signals carry the most
	// current state. Within a generation, skip individual retries
	// (RetryIndex > 0) to avoid duplicates.
	latestGroupRetryIdx := newGroupRetryIdx - 1
	var primarySteps []app.WorkflowStep
	for _, step := range groupSteps {
		if step.GroupRetryIdx != latestGroupRetryIdx || step.RetryIndex > 0 {
			continue
		}
		primarySteps = append(primarySteps, step)
	}

	// Enforce group retry limit, matching the guard in retryGroup().
	groupMaxRetries := executeworkflowstepgroup.GroupMaxRetriesForSteps(primarySteps)
	if newGroupRetryIdx > groupMaxRetries {
		return fmt.Errorf("group retry %d exceeds max retries %d", newGroupRetryIdx, groupMaxRetries)
	}

	// Clone each primary step
	cloneSteps := make([]workflowactivities.CreateFlowStep, 0, len(primarySteps))
	for i, step := range primarySteps {
		// If the signal implements Clone(), use it to produce a clean copy.
		var qs *signaldb.SignalData
		if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
			if cl, ok := step.QueueSignal.Signal.(signal.SignalWithClone); ok {
				defs, cloneErr := cl.Clone(ctx, step.Name)
				if cloneErr != nil {
					return fmt.Errorf("unable to clone signal for retry on step %s: %w", step.Name, cloneErr)
				}
				if len(defs) > 0 {
					// Use the last def: for group retry each step already exists
					// separately, so we want the self-copy (last), not multi-step
					// expansion. Plan Clone() returns [plan] → last=plan. Apply
					// Clone() returns [plan, apply] → last=apply.
					qs = &signaldb.SignalData{Signal: defs[len(defs)-1].Signal}
				}
			} else {
				qs = &signaldb.SignalData{Signal: step.QueueSignal.Signal}
			}
		}

		cloneSteps = append(cloneSteps, workflowactivities.CreateFlowStep{
			FlowID:      s.WorkflowID,
			OwnerID:     step.OwnerID,
			OwnerType:   step.OwnerType,
			Name:        step.Name,
			Signal:      step.Signal,
			QueueSignal: qs,
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
				"is_retry":        true,
				"retry_idx":       newGroupRetryIdx,
				"group_retry_idx": newGroupRetryIdx,
			}),
			Idx:                 maxIdx + 100*(i+1),
			ExecutionType:       step.ExecutionType,
			Metadata:            step.Metadata,
			Retryable:           step.Retryable,
			Skippable:           step.Skippable,
			GroupIdx:            step.GroupIdx,
			GroupRetryIdx:       newGroupRetryIdx,
			StepTargetType:      step.StepTargetType,
			RetryIndex:          0,
			WorkflowStepGroupID: newGroupID,
			Timeout:             step.Timeout,
		})
	}

	if _, err := workflowactivities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, workflowactivities.CreateFlowStepsRequest{
		Steps: cloneSteps,
	}); err != nil {
		return fmt.Errorf("unable to create clone steps: %w", err)
	}

	return nil
}
