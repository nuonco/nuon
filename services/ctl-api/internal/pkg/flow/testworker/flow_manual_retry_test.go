package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestManualRetryOnErroredStep verifies that a user can trigger a retry on a
// failed step via the flow client. After retry:
// - The original step should be marked as discarded
// - A clone step should exist in the same group with RetryIndex=1
// - The clone should execute (and fail again since it uses FailSignal)
// - The workflow should re-enter error state
func (e *FlowTestSuite) TestManualRetryOnErroredStep() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	failSignal := &FailSignal{Reason: "manual retry test"}
	afterSignal := &SuccessSignal{}

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "will-fail", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:   true,
			QueueSignal: &signaldb.SignalData{Signal: failSignal}},
		{Name: "after-fail", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: afterSignal}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	// Find the failed step
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	var failedStep *app.WorkflowStep
	for i := range steps {
		if steps[i].Name == "will-fail" && steps[i].Status.Status == app.StatusError {
			failedStep = &steps[i]
			break
		}
	}
	require.NotNil(e.T(), failedStep, "should find a failed step named 'will-fail'")

	// Trigger manual retry
	resp, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStep.ID,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), resp.Retryable)

	// Wait for the clone to be created and reach a terminal status.
	// We poll steps directly because the workflow StatusError from before the
	// retry fires before the clone has executed.
	var original, clone *app.WorkflowStep
	require.Eventually(e.T(), func() bool {
		steps = e.getStepsByWorkflow(ctx, flw.ID)
		original = nil
		clone = nil
		for i := range steps {
			s := &steps[i]
			if s.GroupIdx != 1 {
				continue
			}
			if s.ID == failedStep.ID {
				original = s
			} else if s.RetryIndex == 1 {
				clone = s
			}
		}
		if clone == nil {
			return false
		}
		// Wait until the clone reaches a terminal status
		return isTerminal(clone.Status.Status)
	}, pollTimeout, pollInterval, "clone step should exist and reach a terminal status")

	require.NotNil(e.T(), original, "original step should still exist")
	require.Equal(e.T(), app.StatusDiscarded, original.Status.Status,
		"original step should be discarded after retry")

	require.NotNil(e.T(), original, "original step should still exist")
	require.Equal(e.T(), app.StatusDiscarded, original.Status.Status,
		"original step should be discarded after retry")

	require.NotNil(e.T(), clone, "clone step should exist with RetryIndex=1")
	require.Equal(e.T(), 1, clone.GroupIdx, "clone should be in the same group")
	require.Equal(e.T(), app.StatusError, clone.Status.Status,
		"clone should have executed and failed (FailSignal always fails)")
	require.Equal(e.T(), "will-fail", clone.Name,
		"clone should have the same name as the original")
}
