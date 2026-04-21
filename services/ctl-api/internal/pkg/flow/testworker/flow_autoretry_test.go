package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestAutoRetryCreatesCloneAndContinues verifies that when a step signal
// implements SignalWithAutoRetry, the step is cloned and the group continues
// executing the clone. After max retries are exhausted, the workflow errors.
func (e *FlowTestSuite) TestAutoRetryCreatesCloneAndContinues() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "auto-retry-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable: true,
			QueueSignal: &signaldb.SignalData{Signal: &AutoRetrySignal{
				FailUntilRetryIndex: 3,
			}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	// The auto-retry signal has MaxRetries=3 and always fails.
	// It should create 3 clones (retry 1, 2, 3) then fail on the 4th attempt.
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	// Verify clones were created — should have original + 3 clones = 4 steps total
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	require.GreaterOrEqual(e.T(), len(steps), 4,
		"expected at least 4 steps (original + 3 retries), got %d", len(steps))

	// All steps should have error or discarded status
	for _, step := range steps {
		require.Contains(e.T(),
			[]app.Status{app.StatusError, app.StatusDiscarded, app.StatusNotAttempted},
			step.Status.Status,
			"step %s has unexpected status %s", step.Name, step.Status.Status)
	}

	// Verify ResultDirective was written on retried steps
	for _, step := range steps[:len(steps)-1] {
		if step.Status.Status == app.StatusError {
			require.Equal(e.T(), "retry", step.ResultDirective,
				"retried step %s should have ResultDirective=retry", step.Name)
		}
	}
}
