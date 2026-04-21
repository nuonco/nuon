package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestAutoRetryEventuallySucceeds verifies that a step that fails twice then
// succeeds on the third attempt (RetryIndex=2) completes the workflow.
func (e *FlowTestSuite) TestAutoRetryEventuallySucceeds() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "countdown-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable: true,
			QueueSignal: &signaldb.SignalData{Signal: &CountdownSignal{
				SucceedAtRetry: 2, // Fails at RetryIndex 0 and 1, succeeds at 2
			}}},
		{Name: "after-retry", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	// Verify there are clone steps (original + 2 retries = 3 steps in group 1)
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	group1Steps := 0
	var successStep *app.WorkflowStep
	for i := range steps {
		if steps[i].GroupIdx == 1 {
			group1Steps++
			if steps[i].Status.Status == app.StatusSuccess {
				successStep = &steps[i]
			}
		}
	}
	require.GreaterOrEqual(e.T(), group1Steps, 3, "expected original + 2 clones in group 1")
	require.NotNil(e.T(), successStep, "one step in group 1 should have succeeded")
	require.Equal(e.T(), 2, successStep.RetryIndex, "successful step should be RetryIndex=2")

	// Group 2 should have executed
	for i := range steps {
		if steps[i].GroupIdx == 2 {
			require.Equal(e.T(), app.StatusSuccess, steps[i].Status.Status,
				"group 2 step should have succeeded after retry")
		}
	}
}

// TestResultDirectiveWrittenOnRetry verifies that auto-retried steps have
// ResultDirective="retry" written to the DB column.
func (e *FlowTestSuite) TestResultDirectiveWrittenOnRetry() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "retry-directive-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable: true,
			QueueSignal: &signaldb.SignalData{Signal: &CountdownSignal{
				SucceedAtRetry: 1, // Fails once, succeeds on retry 1
			}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	steps := e.getStepsByWorkflow(ctx, flw.ID)

	// The original step (RetryIndex=0) should have ResultDirective="retry"
	for _, step := range steps {
		if step.GroupIdx == 1 && step.RetryIndex == 0 {
			require.Equal(e.T(), "retry", step.ResultDirective,
				"original step should have ResultDirective=retry after auto-retry")
		}
	}
}
