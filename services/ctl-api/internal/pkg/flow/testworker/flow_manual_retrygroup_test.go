package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (e *FlowTestSuite) TestManualRetryGroup() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	planSignal := &SuccessSignal{}
	applySignal := &FailSignal{Reason: "manual group retry test"}
	finalizeSignal := &SuccessSignal{}

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

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	var failedStepID string
	for _, step := range steps {
		if step.Name == "g1-apply" && step.Status.Status == app.StatusError {
			failedStepID = step.ID
			break
		}
	}
	require.NotEmpty(e.T(), failedStepID)

	resp, err := e.service.FlowClient.RetryGroup(ctx, &flowclient.RetryGroupRequest{
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
			if step.Name == "g1-apply" && step.GroupRetryIdx == 1 && step.Status.Status == app.StatusError {
				return true
			}
		}
		return false
	}, pollTimeout, pollInterval)

	steps = e.getStepsByWorkflow(ctx, flw.ID)
	g1Generations := make(map[int]int)
	for _, step := range steps {
		if step.GroupIdx == 1 {
			g1Generations[step.GroupRetryIdx]++
		}
	}

	require.Equal(e.T(), 2, g1Generations[0], "original group should have 2 steps")
	require.Equal(e.T(), 2, g1Generations[1], "cloned group should have 2 steps")

	for _, step := range steps {
		if step.GroupIdx == 1 && step.GroupRetryIdx == 0 {
			if step.Name == "g1-plan" {
				require.Equal(e.T(), app.StatusSuccess, step.Status.Status)
			} else {
				require.Equal(e.T(), app.StatusError, step.Status.Status)
				require.True(e.T(), step.Retried)
			}
		}
	}
	e.cancelWorkflow(ctx, flw.ID)
	e.assertTemporalDrained(ctx, flw.ID)
}

func (e *FlowTestSuite) TestManualRetryStepWithRetryGroup() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	planSignal := &SuccessSignal{}
	applySignal := &ManualRetryGroupCountdownSignal{}
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

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

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

	resp, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStepID,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), resp.Retryable)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	steps = e.getStepsByWorkflow(ctx, flw.ID)
	g1Generations := make(map[int]int)
	for _, step := range steps {
		if step.GroupIdx == 1 {
			g1Generations[step.GroupRetryIdx]++
		}
	}
	require.GreaterOrEqual(e.T(), len(g1Generations), 2,
		"expected original and retry group generations, got %v", g1Generations)

	g2Succeeded := false
	for _, step := range steps {
		if step.GroupIdx == 2 && step.Status.Status == app.StatusSuccess {
			g2Succeeded = true
			break
		}
	}
	require.True(e.T(), g2Succeeded, "group 2 finalize should have succeeded")
	e.assertTemporalDrained(ctx, flw.ID)
}

func (e *FlowTestSuite) TestManualRetryAfterAutoBudgetExhausted() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	budgetSignal := &AutoRetryBudgetSignal{}

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "budget-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:   true,
			QueueSignal: &signaldb.SignalData{Signal: budgetSignal}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

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
	require.Equal(e.T(), 1, maxGroupRetryIdx, "auto budget should have produced exactly one retry generation")

	resp, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStepID,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), resp.Retryable)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	steps = e.getStepsByWorkflow(ctx, flw.ID)
	generations := make(map[int]bool)
	for _, step := range steps {
		if step.GroupIdx != 1 {
			continue
		}
		generations[step.GroupRetryIdx] = true
		switch step.GroupRetryIdx {
		case 0:
			require.Equal(e.T(), app.StatusError, step.Status.Status)
			require.True(e.T(), step.Retried)
		case 1:
			require.Equal(e.T(), app.StatusDiscarded, step.Status.Status)
			require.True(e.T(), step.Retried)
		default:
			require.Equal(e.T(), app.StatusSuccess, step.Status.Status,
				"manual retry generation should have succeeded")
		}
	}
	require.Len(e.T(), generations, 3, "expected three group generations (original, auto retry, manual retry)")
	e.assertTemporalDrained(ctx, flw.ID)
}
