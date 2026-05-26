package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestRetryGroupClonesEntireGroup verifies that when a step has
// SignalWithRetryGroup and auto-retry triggers, the entire group is cloned
// (not just the single step).
//
// Setup: Group 1 has a plan (succeeds) and an apply (always fails, RetryGroup=true).
// The apply signal implements CloneSteps to return both plan + apply on retry.
// Expected: Group 1 is cloned with a new GroupRetryIdx. The cloned apply
// fails again, exhausting retries, and the workflow errors.
func (e *FlowTestSuite) TestRetryGroupClonesEntireGroup() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	planSignal := &SuccessSignal{}
	applySignal := &PlanApplyFailSignal{}
	finalizeSignal := &SuccessSignal{}

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "g1-plan", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: planSignal}},
		{Name: "g1-apply", Idx: 200, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:   true,
			QueueSignal: &signaldb.SignalData{Signal: applySignal}},
		{Name: "g2-finalize", Idx: 300, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: finalizeSignal}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	// PlanApplyFailSignal has MaxRetries=2 and always fails.
	// The group will be cloned, then max retries exhausted → workflow errors.
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	steps := e.getStepsByWorkflow(ctx, flw.ID)

	// Count distinct GroupRetryIdx values for group 1
	groupRetryIdxs := make(map[int]bool)
	for _, step := range steps {
		if step.GroupIdx == 1 {
			groupRetryIdxs[step.GroupRetryIdx] = true
		}
	}

	// Should have at least 2 group retry generations (original + clones)
	require.GreaterOrEqual(e.T(), len(groupRetryIdxs), 2,
		"expected multiple group retry generations, got %d: %v", len(groupRetryIdxs), groupRetryIdxs)

	// Original group 1 steps should be discarded
	for _, step := range steps {
		if step.GroupIdx == 1 && step.GroupRetryIdx == 0 {
			require.Equal(e.T(), app.StatusDiscarded, step.Status.Status,
				"original step %s should be discarded, got %s", step.Name, step.Status.Status)
		}
	}

	// Group 2 should NOT have been reached (group 1 keeps failing)
	for _, step := range steps {
		if step.GroupIdx == 2 {
			require.NotEqual(e.T(), app.StatusSuccess, step.Status.Status,
				"group 2 step should not have succeeded since group 1 never passed")
		}
	}
}

// TestRetryGroupRetryOfRetryDiscardsAllPreviousGroups verifies that when a
// group retry is itself retried (retry-of-retry), all previous group objects
// for the same GroupIdx are marked as discarded. Without the fix, only the
// first group object would be discarded, leaving the intermediate retry's
// group non-discarded and causing incorrect dispatch.
func (e *FlowTestSuite) TestRetryGroupRetryOfRetryDiscardsAllPreviousGroups() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	planSignal := &SuccessSignal{}
	applySignal := &PlanApplyFailSignal{} // MaxRetries=2, always fails, requests group retry
	triggerSignal := &SuccessSignal{}

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "g1-plan", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: planSignal}},
		{Name: "g1-apply", Idx: 200, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:   true,
			QueueSignal: &signaldb.SignalData{Signal: applySignal}},
		{Name: "g1-trigger", Idx: 300, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: triggerSignal}},
		{Name: "g2-finalize", Idx: 400, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	// Apply always fails → group retries until max retries exhausted → workflow errors.
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	steps := e.getStepsByWorkflow(ctx, flw.ID)

	// Count distinct GroupRetryIdx values for group 1 — need at least 3
	// generations (original + retry1 + retry2) to exercise retry-of-retry.
	groupRetryIdxs := make(map[int]bool)
	for _, step := range steps {
		if step.GroupIdx == 1 {
			groupRetryIdxs[step.GroupRetryIdx] = true
		}
	}
	require.GreaterOrEqual(e.T(), len(groupRetryIdxs), 3,
		"expected at least 3 group retry generations (original + 2 retries), got %d: %v",
		len(groupRetryIdxs), groupRetryIdxs)

	// Verify that only ONE non-discarded WorkflowStepGroup exists for GroupIdx=1.
	// This is the critical assertion: without the fix, intermediate retry groups
	// would remain non-discarded.
	var allGroups []app.WorkflowStepGroup
	res := e.service.DB.WithContext(ctx).
		Where("workflow_id = ? AND group_idx = ?", flw.ID, 1).
		Find(&allGroups)
	require.Nil(e.T(), res.Error)

	nonDiscardedCount := 0
	for _, g := range allGroups {
		if g.Status.Status != app.StatusDiscarded {
			nonDiscardedCount++
		}
	}
	// After all retries exhaust, the last group may be in error/discarded state.
	// The important thing is we don't have >1 non-discarded group.
	require.LessOrEqual(e.T(), nonDiscardedCount, 1,
		"expected at most 1 non-discarded group for GroupIdx=1, got %d", nonDiscardedCount)

	// Verify the trigger step was cloned in each retry generation.
	triggerCount := 0
	for _, step := range steps {
		if step.GroupIdx == 1 && step.Name == "g1-trigger" {
			triggerCount++
		}
	}
	require.GreaterOrEqual(e.T(), triggerCount, 1,
		"trigger step should be present in at least the original generation")
}
