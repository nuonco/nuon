package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (e *FlowTestSuite) TestSingleStepSuccess() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	e.phase("seed")

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "only-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})
	e.phase("fixtures")

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.phase("enqueue")

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	e.phase("db-success")
	require.Eventually(e.T(), func() bool {
		queueSignal := e.getLatestQueueSignal(ctx, flw.ID, "install_workflows", executeflow.SignalType)
		return queueSignal.Status.Status == app.StatusSuccess
	}, testResidentIdleTimeout, pollInterval, "resident success should not wait for the idle timeout")

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	require.Len(e.T(), steps, 1)
	require.Equal(e.T(), app.StatusSuccess, steps[0].Status.Status)
	e.phase("assertions")

	e.assertTemporalDrained(ctx, flw.ID)
	e.phase("drain")
}

func (e *FlowTestSuite) TestNoStepsNoSignalErrors() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	stepQueue := e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-steps")
	e.createTestQueue(ctx, ownerID, ownerType, "install-signals")

	flw := app.Workflow{
		OwnerID:   ownerID,
		OwnerType: ownerType,
		Type:      "test_flow",
		Status:    app.NewCompositeStatus(ctx, app.StatusPending),
	}
	res := e.service.DB.WithContext(ctx).Create(&flw)
	require.Nil(e.T(), res.Error)

	e.enqueueFlow(ctx, stepQueue.ID, &flw, ownerID, ownerType)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	e.assertTemporalDrained(ctx, flw.ID)
}
