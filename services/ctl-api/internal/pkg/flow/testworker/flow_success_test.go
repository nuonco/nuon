package testworker

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// setupFlowTest creates the queues, workflow, and steps needed for a flow test.
// Returns the workflow and the step queue ID for enqueuing the execute-flow signal.
func (e *FlowTestSuite) setupFlowTest(ctx context.Context, ownerID, ownerType string, steps []app.WorkflowStep) (*app.Workflow, string) {
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

	e.createTestSteps(ctx, &flw, steps)

	return &flw, stepQueue.ID
}

// enqueueFlow dispatches the execute-flow signal to start the workflow.
func (e *FlowTestSuite) enqueueFlow(ctx context.Context, queueID string, flw *app.Workflow, ownerID, ownerType string) {
	resp, err := e.service.QueueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queueID,
		Signal: &executeflow.Signal{
			WorkflowID:          flw.ID,
			StepQueueName:       "install-workflow-steps",
			StepTargetQueueName: "install-signals",
			OwnerID:             ownerID,
			OwnerType:           ownerType,
		},
		// Set owner so the flow client can find this queue signal via
		// findQueueSignalByOwner(workflowID, "install_workflows", ...).
		OwnerID:   flw.ID,
		OwnerType: "install_workflows",
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), resp)
}

// TestSequentialGroupSuccess verifies that a workflow with multiple groups
// executes all steps sequentially and completes with StatusSuccess.
func (e *FlowTestSuite) TestSequentialGroupSuccess() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "g1-step1", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "g1-step2", Idx: 200, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "g2-step1", Idx: 300, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "g2-step2", Idx: 400, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	// Verify all steps completed
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		require.Equal(e.T(), app.StatusSuccess, step.Status.Status,
			"step %s should be success, got %s", step.Name, step.Status.Status)
	}
}
