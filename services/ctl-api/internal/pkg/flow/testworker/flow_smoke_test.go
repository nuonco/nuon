package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestSingleStepSuccess is the simplest possible flow test: one group, one step.
func (e *FlowTestSuite) TestSingleStepSuccess() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "only-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	require.Len(e.T(), steps, 1)
	require.Equal(e.T(), app.StatusSuccess, steps[0].Status.Status)
}

// TestNoStepsNoSignalErrors verifies that a workflow with no pre-created steps
// and no GenerateStepsSignal fails with a clear error.
func (e *FlowTestSuite) TestNoStepsNoSignalErrors() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-steps")
	e.createTestQueue(ctx, ownerID, ownerType, "install-signals")

	stepQueue := e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-steps")
	e.createTestQueue(ctx, ownerID, ownerType, "install-signals")

	// Create workflow with no steps and no GenerateStepsSignal
	flw := app.Workflow{
		OwnerID:   ownerID,
		OwnerType: ownerType,
		Type:      "test_flow",
		Status:    app.NewCompositeStatus(ctx, app.StatusPending),
	}
	res := e.service.DB.WithContext(ctx).Create(&flw)
	require.Nil(e.T(), res.Error)

	e.enqueueFlow(ctx, stepQueue.ID, &flw, ownerID, ownerType)

	// Should error because there are no steps and no way to generate them
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)
}
