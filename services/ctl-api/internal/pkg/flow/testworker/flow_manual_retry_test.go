package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

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

	steps = e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		require.NotEqual(e.T(), app.StatusSuccess, step.Status.Status,
			"step %s (group %d) should not have succeeded", step.Name, step.GroupIdx)
	}
	e.cancelWorkflow(ctx, flw.ID)
	e.assertTemporalDrained(ctx, flw.ID)
}
