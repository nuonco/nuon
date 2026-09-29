package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (e *FlowTestSuite) TestResumeStartsAtCorrectGroup() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

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

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		if step.Name == "g1-step" {
			require.Equal(e.T(), app.StatusSuccess, step.Status.Status,
				"group 1 step should have succeeded")
		}
	}

	var failedStepID string
	for _, step := range steps {
		if step.Name == "g2-step" && step.Status.Status == app.StatusError {
			failedStepID = step.ID
			break
		}
	}
	require.NotEmpty(e.T(), failedStepID)

	resp, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStepID,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), resp.Retryable)

	require.Eventually(e.T(), func() bool {
		steps, err := e.tryStepsByWorkflow(ctx, flw.ID)
		if err != nil {
			return false
		}
		for _, step := range steps {
			if step.GroupIdx == 2 && step.RetryIndex == 1 && step.Status.Status == app.StatusError {
				return true
			}
		}
		return false
	}, pollTimeout, pollInterval)

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

	g2StepCount := 0
	for _, step := range steps {
		if step.GroupIdx == 2 {
			g2StepCount++
		}
	}
	require.GreaterOrEqual(e.T(), g2StepCount, 2, "group 2 should have original + retry clone")

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	e.cancelWorkflow(ctx, flw.ID)
	e.assertTemporalDrained(ctx, flw.ID)
}

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

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	var failedStepID string
	for _, step := range steps {
		if step.Name == "will-fail" && step.Status.Status == app.StatusError {
			failedStepID = step.ID
			break
		}
	}
	require.NotEmpty(e.T(), failedStepID)

	skipResp, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStepID,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), skipResp.Skippable)

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

	failedCount := 0
	for _, step := range steps {
		if step.Name == "will-fail" {
			failedCount++
		}
	}
	require.Equal(e.T(), 1, failedCount, "skip must not clone the skipped step")

	e.assertTemporalDrained(ctx, flw.ID)
}

func (e *FlowTestSuite) TestSkipParkedStep() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "will-fail", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:   true,
			Skippable:   true,
			QueueSignal: &signaldb.SignalData{Signal: &ManualRetryGroupCountdownSignal{}}},
		{Name: "after-skip", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	var failedStepID string
	for _, step := range steps {
		if step.Name == "will-fail" && step.Status.Status == app.StatusError {
			failedStepID = step.ID
			break
		}
	}
	require.NotEmpty(e.T(), failedStepID)

	skipResp, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStepID,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), skipResp.Skippable)

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

	e.assertTemporalDrained(ctx, flw.ID)
}
