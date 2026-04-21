package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestManualRetryGroup verifies that a user can trigger a group-level retry
// via the flow client, causing the entire group to be cloned and re-executed.
func (e *FlowTestSuite) TestManualRetryGroup() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	planSignal := &SuccessSignal{}
	applySignal := &FailSignal{Reason: "manual group retry test"}
	finalizeSignal := &SuccessSignal{}

	// Group 1: plan succeeds, apply fails
	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "g1-plan", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: planSignal}},
		{Name: "g1-apply", Idx: 200, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:   true,
			QueueSignal: &signaldb.SignalData{Signal: applySignal}},
		{Name: "g2-step", Idx: 300, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: finalizeSignal}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	// Find the failed step
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	var failedStepID string
	for _, step := range steps {
		if step.Name == "g1-apply" && step.Status.Status == app.StatusError {
			failedStepID = step.ID
			break
		}
	}
	require.NotEmpty(e.T(), failedStepID)

	// Trigger group retry via the flow client
	resp, err := e.service.FlowClient.RetryGroup(ctx, &flowclient.RetryGroupRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStepID,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), resp.Retryable)

	// The workflow resumes. The group will be re-dispatched. The cloned apply
	// step will also fail (same FailSignal), so the workflow errors again.
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	// Verify the entire group was cloned — both plan and apply should have clones
	steps = e.getStepsByWorkflow(ctx, flw.ID)
	g1Generations := make(map[int]int) // GroupRetryIdx -> count
	for _, step := range steps {
		if step.GroupIdx == 1 {
			g1Generations[step.GroupRetryIdx]++
		}
	}

	// Original generation (GroupRetryIdx=0) should have 2 steps
	require.Equal(e.T(), 2, g1Generations[0], "original group should have 2 steps")
	// Clone generation (GroupRetryIdx=1) should also have 2 steps
	require.Equal(e.T(), 2, g1Generations[1], "cloned group should have 2 steps")

	// Original steps should be discarded
	for _, step := range steps {
		if step.GroupIdx == 1 && step.GroupRetryIdx == 0 {
			require.Equal(e.T(), app.StatusDiscarded, step.Status.Status,
				"original step %s should be discarded", step.Name)
		}
	}
}

// TestManualRetryStepWithRetryGroup verifies that calling RetryStep on a step
// whose signal implements RetryGroup correctly clones the entire group (not just
// the single step). The test uses ManualRetryGroupCountdownSignal which:
//   - Auto-retries once (group clone, generation 1) and fails again
//   - Exhausts max retries → workflow errors
//   - Manual RetryStep triggers another group clone (generation 2)
//   - Signal sees GroupRetryCount=2 and succeeds
func (e *FlowTestSuite) TestManualRetryStepWithRetryGroup() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	planSignal := &SuccessSignal{}
	applySignal := &ManualRetryGroupCountdownSignal{}
	finalizeSignal := &SuccessSignal{}

	// Group 1: plan (succeeds) + apply (countdown that needs 2 group retries)
	// Group 2: finalize (succeeds once group 1 passes)
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

	// Phase 1: auto-retry produces generation 1, which also fails → workflow errors
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	// Find the most recent failed apply step (highest GroupRetryIdx)
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	var failedStepID string
	maxGroupRetryIdx := -1
	for _, step := range steps {
		if step.GroupIdx == 1 && step.Status.Status == app.StatusError && step.GroupRetryIdx > maxGroupRetryIdx {
			failedStepID = step.ID
			maxGroupRetryIdx = step.GroupRetryIdx
		}
	}
	require.NotEmpty(e.T(), failedStepID, "should have a failed apply step")

	// Phase 2: manual RetryStep (NOT RetryGroup!) — the new code should detect
	// RetryGroup() on the signal and clone the entire group.
	resp, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStepID,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), resp.Retryable)

	// Phase 3: the cloned group (generation 2) executes — apply sees GroupRetryCount=2 → succeeds
	// Group 2 finalize also succeeds → workflow completes
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	// Verify: at least 3 group generations for group 1
	steps = e.getStepsByWorkflow(ctx, flw.ID)
	g1Generations := make(map[int]int)
	for _, step := range steps {
		if step.GroupIdx == 1 {
			g1Generations[step.GroupRetryIdx]++
		}
	}
	require.GreaterOrEqual(e.T(), len(g1Generations), 3,
		"expected at least 3 group retry generations (0, 1, 2), got %v", g1Generations)

	// Verify group 2 was reached and succeeded
	g2Succeeded := false
	for _, step := range steps {
		if step.GroupIdx == 2 && step.Status.Status == app.StatusSuccess {
			g2Succeeded = true
			break
		}
	}
	require.True(e.T(), g2Succeeded, "group 2 finalize should have succeeded")
}
