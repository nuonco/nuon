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
