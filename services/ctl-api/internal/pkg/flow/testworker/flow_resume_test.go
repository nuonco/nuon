package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestResumeStartsAtCorrectGroup verifies that when group 1 succeeds and
// group 2 fails, a retry resumes at group 2 (not group 1).
// This validates the Bug 1 fix (resumeStartIdx = findGroupPositionForStep).
func (e *FlowTestSuite) TestResumeStartsAtCorrectGroup() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	// Group 1 succeeds, group 2 has a countdown that fails first then succeeds.
	// But since we need manual retry to test resume position, use FailSignal
	// in group 2.
	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "g1-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "g2-step", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:   true,
			QueueSignal: &signaldb.SignalData{Signal: &FailSignal{Reason: "first attempt"}}},
		{Name: "g3-step", Idx: 300, GroupIdx: 3, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	// Wait for workflow to error at group 2
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	// Verify group 1 succeeded
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		if step.Name == "g1-step" {
			require.Equal(e.T(), app.StatusSuccess, step.Status.Status,
				"group 1 step should have succeeded")
		}
	}

	// Find the failed step for retry
	var failedStepID string
	for _, step := range steps {
		if step.Name == "g2-step" && step.Status.Status == app.StatusError {
			failedStepID = step.ID
			break
		}
	}
	require.NotEmpty(e.T(), failedStepID)

	// Retry the failed step
	resp, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStepID,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), resp.Retryable)

	// The workflow will resume. The clone of g2-step will also fail (same FailSignal).
	// The workflow should error again.
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	// Key assertion: group 1 step should NOT have been re-executed.
	// It should still have the same single step with StatusSuccess.
	steps = e.getStepsByWorkflow(ctx, flw.ID)
	g1StepCount := 0
	for _, step := range steps {
		if step.GroupIdx == 1 {
			g1StepCount++
			require.Equal(e.T(), app.StatusSuccess, step.Status.Status,
				"group 1 step should still be success (not re-executed)")
		}
	}
	require.Equal(e.T(), 1, g1StepCount, "group 1 should have exactly 1 step (not re-run)")

	// Group 2 should have the original + clone
	g2StepCount := 0
	for _, step := range steps {
		if step.GroupIdx == 2 {
			g2StepCount++
		}
	}
	require.GreaterOrEqual(e.T(), g2StepCount, 2, "group 2 should have original + retry clone")
}

// TestSkipErroredStep verifies that skipping a failed step causes the workflow
// to resume past it.
func (e *FlowTestSuite) TestSkipErroredStep() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "will-fail", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:   true,
			Skippable:   true,
			QueueSignal: &signaldb.SignalData{Signal: &FailSignal{Reason: "skip test"}}},
		{Name: "after-skip", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	// Find failed step
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	var failedStepID string
	for _, step := range steps {
		if step.Name == "will-fail" && step.Status.Status == app.StatusError {
			failedStepID = step.ID
			break
		}
	}
	require.NotEmpty(e.T(), failedStepID)

	// Skip it
	skipResp, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStepID,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), skipResp.Skippable)

	// Workflow should resume and succeed (group 2 runs)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	steps = e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		if step.Name == "will-fail" {
			require.Equal(e.T(), app.StatusUserSkipped, step.Status.Status,
				"skipped step should have user-skipped status")
		}
		if step.Name == "after-skip" {
			require.Equal(e.T(), app.StatusSuccess, step.Status.Status,
				"step after skip should have succeeded")
		}
	}
}
