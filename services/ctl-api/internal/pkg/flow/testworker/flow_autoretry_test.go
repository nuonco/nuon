package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

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

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	require.GreaterOrEqual(e.T(), len(steps), 4,
		"expected at least 4 steps (original + 3 retries), got %d", len(steps))

	for _, step := range steps {
		require.Contains(e.T(),
			[]app.Status{app.StatusError, app.StatusDiscarded, app.StatusNotAttempted},
			step.Status.Status,
			"step %s has unexpected status %s", step.Name, step.Status.Status)
	}

	for _, step := range steps[:len(steps)-1] {
		if step.Status.Status == app.StatusError {
			require.Equal(e.T(), "retry", step.ResultDirective,
				"retried step %s should have ResultDirective=retry", step.Name)
		}
	}
	e.assertTemporalDrained(ctx, flw.ID)
}
