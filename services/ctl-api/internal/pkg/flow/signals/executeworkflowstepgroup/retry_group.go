package executeworkflowstepgroup

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// retryGroup clones all steps in the group for retry. This is the core RetryGroup
// logic — it lives on the group signal, not on the step signal.
//
// 1. Fetches all steps in the group
// 2. Guards against parallel groups (retry-group not supported)
// 3. Marks all steps as Discarded
// 4. Clones each step (respecting SignalWithCloneSteps)
// 5. Increments GroupRetryIdx on all clones
func (s *Signal) retryGroup(ctx workflow.Context, l *zap.Logger) error {
	if s.Parallel {
		return errors.New("retry-group is not supported for parallel groups")
	}

	steps, err := s.getGroupSteps(ctx)
	if err != nil {
		return err
	}

	if len(steps) == 0 {
		return errors.New("no steps found in group to retry")
	}

	// Determine the new GroupRetryIdx and base Idx for clones
	maxIdx := 0
	newGroupRetryIdx := 0
	for _, step := range steps {
		if step.Idx > maxIdx {
			maxIdx = step.Idx
		}
		if step.GroupRetryIdx >= newGroupRetryIdx {
			newGroupRetryIdx = step.GroupRetryIdx + 1
		}
	}

	l.Debug("retrying group",
		zap.Int("group_idx", s.GroupIdx),
		zap.Int("step_count", len(steps)),
		zap.Int("new_group_retry_idx", newGroupRetryIdx))

	// Mark ALL steps in the group as retried. Keep the existing status so
	// errors still show correctly in the dashboard — just set the retried
	// boolean and add group_retry metadata.
	for _, step := range steps {
		if step.Status.Status == app.StatusDiscarded {
			continue
		}
		_ = activities.AwaitPkgWorkflowsFlowUpdateFlowStepRetried(ctx, activities.UpdateFlowStepRetriedRequest{
			StepID: step.ID,
		})
	}

	// Clone from the latest group retry generation (not always gen 0).
	// On retry-of-retry the latest generation's signals carry the most
	// current state. Within a generation, skip individual retries
	// (RetryIndex > 0) to avoid duplicates.
	latestGroupRetryIdx := newGroupRetryIdx - 1
	var stepsToClone []app.WorkflowStep
	for _, step := range steps {
		if step.GroupRetryIdx != latestGroupRetryIdx || step.RetryIndex > 0 {
			continue
		}
		stepsToClone = append(stepsToClone, step)
	}

	// Enforce per-step max retries at the group level. The group retry count
	// is bounded by the minimum MaxRetries across all signals in the group
	// (including signals produced by CloneSteps). For example, if plan has
	// MaxRetries=3 and apply has MaxRetries=5, the group stops after 3.
	groupMaxRetries := GroupMaxRetriesForSteps(stepsToClone)
	if newGroupRetryIdx > groupMaxRetries {
		l.Warn("group retry exceeds per-step max retries",
			zap.Int("new_group_retry_idx", newGroupRetryIdx),
			zap.Int("group_max_retries", groupMaxRetries))
		return fmt.Errorf("group retry %d exceeds max retries %d", newGroupRetryIdx, groupMaxRetries)
	}

	// Clone each step from the latest generation. CloneSteps expansion was
	// already applied during initial workflow creation, so each generation
	// already represents the full set (e.g. plan + apply as separate steps).
	cloneSteps := make([]activities.CreateFlowStep, 0, len(stepsToClone))
	for i, step := range stepsToClone {
		// If the signal implements Clone(), use it to produce clean copies.
		// For group retry we always produce a single step per original (not
		// multi-step expansion) because the group already has the full set.
		cloneQueueSignal := step.QueueSignal
		if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
			if cl, ok := step.QueueSignal.Signal.(signal.SignalWithClone); ok {
				defs, cloneErr := cl.Clone(ctx, step.Name)
				if cloneErr != nil {
					return errors.Wrapf(cloneErr, "unable to clone signal for retry on step %s", step.Name)
				}
				if len(defs) > 0 {
					// Use the last def: for group retry each step already exists
					// separately, so we want the self-copy (last), not multi-step
					// expansion. Plan Clone() returns [plan] → last=plan. Apply
					// Clone() returns [plan, apply] → last=apply.
					cloneQueueSignal = &signaldb.SignalData{Signal: defs[len(defs)-1].Signal}
				}
			}
		}

		cloneSteps = append(cloneSteps, activities.CreateFlowStep{
			FlowID:      s.WorkflowID,
			OwnerID:     step.OwnerID,
			OwnerType:   step.OwnerType,
			Name:        step.Name,
			Signal:      step.Signal,
			QueueSignal: cloneQueueSignal,
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
				"is_retry":        true,
				"retry_idx":       newGroupRetryIdx,
				"group_retry_idx": newGroupRetryIdx,
				"retry_type":      "auto",
			}),
			Idx:            maxIdx + 100 + i,
			ExecutionType:  step.ExecutionType,
			Metadata:       step.Metadata,
			Retryable:      step.Retryable,
			Skippable:      step.Skippable,
			GroupIdx:       step.GroupIdx,
			GroupRetryIdx:  newGroupRetryIdx,
			StepTargetType: step.StepTargetType,
			RetryIndex:     0,
			Timeout:        step.Timeout,
		})
	}

	if len(cloneSteps) > 0 {
		if _, err := activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, activities.CreateFlowStepsRequest{
			Steps: cloneSteps,
		}); err != nil {
			return errors.Wrap(err, "unable to create retry group clone steps")
		}
	}

	return nil
}

// GroupMaxRetriesForSteps returns the minimum MaxRetries across all signals
// in the given steps, including signals produced by Clone(). Falls back to
// signal.DefaultMaxRetries for signals that don't implement SignalWithMaxRetries.
func GroupMaxRetriesForSteps(steps []app.WorkflowStep) int {
	minRetries := signal.DefaultMaxRetries

	for _, step := range steps {
		if step.QueueSignal == nil || step.QueueSignal.Signal == nil {
			continue
		}
		sig := step.QueueSignal.Signal

		// Check the step's own signal.
		checkMax(sig, &minRetries)
	}

	return minRetries
}

func checkMax(sig signal.Signal, minRetries *int) {
	mr, ok := sig.(signal.SignalWithMaxRetries)
	if !ok {
		return
	}
	if v := mr.MaxRetries(); v < *minRetries {
		*minRetries = v
	}
}
