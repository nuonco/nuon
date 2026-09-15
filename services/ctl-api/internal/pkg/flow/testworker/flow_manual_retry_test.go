package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestManualRetryOnErroredStep pins the manual-retry contract end to end.
//
// Park: ManualRetrySignal sets MaxAutoRetries=0, so the first failure parks
// instead of auto-retrying — process_errors.go writes the StepAwaitManualRetry
// directive and the workflow stops at StatusFailedPendingRetry while the step
// handler blocks on the retry update.
//
// Retry: RetryStep reaches createStepRetryHandler, which marks the original
// step retried=true and StatusDiscarded (metadata retry_type=manual) and
// creates a clone row in the same group with RetryIndex=1. The original keeps
// an audit trail as Discarded; the clone is the row that re-executes.
//
// Outcome: a manual retry that succeeds means the original failure was
// transient. The clone reaches StatusSuccess and the workflow continues to
// StatusSuccess. The failure branch (a clone that can never succeed) is
// pinned separately by TestManualRetryOnAlwaysFailingStep.
//
// Note the asymmetry with group retries: cloneGroupForRetry keeps the original
// steps' StatusError and only sets retried=true, because the whole group's
// error must stay visible in the dashboard.
func (e *FlowTestSuite) TestManualRetryOnErroredStep() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	failSignal := &ManualRetrySignal{}
	afterSignal := &SuccessSignal{}

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "will-fail", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:   true,
			QueueSignal: &signaldb.SignalData{Signal: failSignal}},
		{Name: "after-fail", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: afterSignal}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

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
		fetched, err := e.tryStepsByWorkflow(ctx, flw.ID)
		if err != nil {
			return false
		}
		steps = fetched
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

	require.NotNil(e.T(), clone, "clone step should exist with RetryIndex=1")
	require.Equal(e.T(), 1, clone.GroupIdx, "clone should be in the same group")
	require.Equal(e.T(), app.StatusSuccess, clone.Status.Status,
		"clone should have executed successfully")
	require.Equal(e.T(), "will-fail", clone.Name,
		"clone should have the same name as the original")
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	e.assertTemporalDrained(ctx, flw.ID)
}

// TestManualRetryOnAlwaysFailingStep pins the failure branch of the
// manual-retry contract: a step whose signal can never succeed must stay
// failed after a manual retry. FailSignal implements no retry interfaces, so
// handleStepError marks the step failed without any retry budget — the clone's
// second failure is terminal (step and workflow StatusError, no second park),
// the workflow must not be resurrected into a retry loop, and the downstream
// group must never run. Together with TestManualRetryOnErroredStep this pins
// both directions: transient failure → clone succeeds, persistent failure →
// no step ever reaches StatusSuccess.
func (e *FlowTestSuite) TestManualRetryOnAlwaysFailingStep() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	failSignal := &FailSignal{Reason: "always-failing retry test"}
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

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	var failedStep *app.WorkflowStep
	for i := range steps {
		if steps[i].Name == "will-fail" && steps[i].Status.Status == app.StatusError {
			failedStep = &steps[i]
			break
		}
	}
	require.NotNil(e.T(), failedStep, "should find a failed step named 'will-fail'")

	resp, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStep.ID,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), resp.Retryable)

	// The clone re-executes and fails again.
	var clone *app.WorkflowStep
	require.Eventually(e.T(), func() bool {
		fetched, err := e.tryStepsByWorkflow(ctx, flw.ID)
		if err != nil {
			return false
		}
		clone = nil
		for i := range fetched {
			s := &fetched[i]
			if s.GroupIdx == 1 && s.RetryIndex == 1 && s.Name == "will-fail" {
				clone = s
			}
		}
		return clone != nil && isTerminal(clone.Status.Status)
	}, pollTimeout, pollInterval, "clone step should exist and reach a terminal status")

	require.Equal(e.T(), app.StatusError, clone.Status.Status,
		"clone of an always-failing step should fail again")
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	// The downstream group must never have run.
	steps = e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		require.NotEqual(e.T(), app.StatusSuccess, step.Status.Status,
			"step %s (group %d) should not have succeeded", step.Name, step.GroupIdx)
	}
	e.cancelWorkflow(ctx, flw.ID)
	e.assertTemporalDrained(ctx, flw.ID)
}
