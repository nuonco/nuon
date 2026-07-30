package testworker

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateworkflowsteps"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// setupFlowTest creates the queues, workflow, and steps needed for a flow test.
// Returns the workflow and the step queue ID for enqueuing the execute-flow signal.
func (e *FlowTestSuite) setupFlowTest(ctx context.Context, ownerID, ownerType string, steps []app.WorkflowStep) (*app.Workflow, string) {
	stepQueue := e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-steps")
	e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-step-groups")
	e.createTestQueue(ctx, ownerID, ownerType, "install-signals")
	e.createTestQueue(ctx, ownerID, ownerType, "install-generate-steps")

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

func (e *FlowTestSuite) setupGroupedFlowTest(ctx context.Context, ownerID, ownerType string, steps []app.WorkflowStep) (*app.Workflow, string) {
	stepQueue := e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-steps")
	e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-step-groups")
	e.createTestQueue(ctx, ownerID, ownerType, "install-signals")
	e.createTestQueue(ctx, ownerID, ownerType, "install-generate-steps")

	flw := app.Workflow{
		OwnerID:   ownerID,
		OwnerType: ownerType,
		Type:      "test_flow",
		Status:    app.NewCompositeStatus(ctx, app.StatusPending),
	}
	res := e.service.DB.WithContext(ctx).Create(&flw)
	require.Nil(e.T(), res.Error)

	groups := make(map[int]string)
	for i := range steps {
		groupID, ok := groups[steps[i].GroupIdx]
		if !ok {
			group := app.WorkflowStepGroup{
				WorkflowID: flw.ID,
				GroupIdx:   steps[i].GroupIdx,
				Parallel:   steps[i].GroupParallel,
				Status:     app.NewCompositeStatus(ctx, app.StatusPending),
			}
			res := e.service.DB.WithContext(ctx).Create(&group)
			require.Nil(e.T(), res.Error)
			groupID = group.ID
			groups[steps[i].GroupIdx] = groupID
		}
		steps[i].WorkflowStepGroupID = groupID
	}

	e.createTestSteps(ctx, &flw, steps)
	return &flw, stepQueue.ID
}

// enqueueFlow dispatches the execute-flow signal to start the workflow.
func (e *FlowTestSuite) enqueueFlow(ctx context.Context, queueID string, flw *app.Workflow, ownerID, ownerType string) {
	resp, err := e.service.QueueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queueID,
		Signal: &executeflow.Signal{
			WorkflowID:             flw.ID,
			StepGroupQueueName:     "install-workflow-step-groups",
			StepQueueName:          "install-workflow-steps",
			StepTargetQueueName:    "install-signals",
			GenerateStepsQueueName: "install-generate-steps",
			OwnerID:                ownerID,
			OwnerType:              ownerType,
		},
		// Set owner so the flow client can find this queue signal via
		// findQueueSignalByOwner(workflowID, "install_workflows", ...).
		OwnerID:   flw.ID,
		OwnerType: "install_workflows",
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), resp)
}

func (e *FlowTestSuite) enqueueResidentFlow(ctx context.Context, queueID string, flw *app.Workflow, ownerID, ownerType string, idleTimeout time.Duration) {
	resp, err := e.service.QueueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queueID,
		Signal: &executeflow.Signal{
			WorkflowID:             flw.ID,
			StepGroupQueueName:     "install-workflow-step-groups",
			StepQueueName:          "install-workflow-steps",
			StepTargetQueueName:    "install-signals",
			GenerateStepsQueueName: "install-generate-steps",
			OwnerID:                ownerID,
			OwnerType:              ownerType,
			Resident:               true,
			ResidentIdleTimeout:    idleTimeout,
		},
		OwnerID:   flw.ID,
		OwnerType: "install_workflows",
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), resp)
}

func (e *FlowTestSuite) enqueueGroupedLegacyFlow(ctx context.Context, queueID string, flw *app.Workflow, ownerID, ownerType string) {
	resp, err := e.service.QueueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queueID,
		Signal: &executeflow.Signal{
			WorkflowID:             flw.ID,
			StepGroupQueueName:     "install-workflow-step-groups",
			StepQueueName:          "install-workflow-steps",
			StepTargetQueueName:    "install-signals",
			GenerateStepsQueueName: "install-generate-steps",
			OwnerID:                ownerID,
			OwnerType:              ownerType,
		},
		OwnerID:   flw.ID,
		OwnerType: "install_workflows",
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), resp)
}

func (e *FlowTestSuite) TestNewExecuteFlowSignalPersistsResidentDefault() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{{
		Name:          "resident-default",
		Idx:           100,
		GroupIdx:      1,
		ExecutionType: app.WorkflowStepExecutionTypeSystem,
		QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
	}})

	resp, err := e.service.QueueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID:   queueID,
		Signal:    executeflow.NewSignal(flw.ID),
		OwnerID:   flw.ID,
		OwnerType: "install_workflows",
	})
	require.NoError(e.T(), err)
	require.NotNil(e.T(), resp)

	queueSignal := e.getLatestQueueSignal(ctx, flw.ID, "install_workflows", executeflow.SignalType)
	persistedSignal, ok := queueSignal.Signal.Signal.(*executeflow.Signal)
	require.True(e.T(), ok)
	require.True(e.T(), persistedSignal.Resident)
}

func (e *FlowTestSuite) TestResidentGeneratedWorkflowExecutesSteps() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	workflowType := app.WorkflowType("test_resident_generated_flow")

	generateworkflowsteps.RegisterGenerators(ownerType, func() map[app.WorkflowType]flow.WorkflowStepGenerator {
		return map[app.WorkflowType]flow.WorkflowStepGenerator{
			workflowType: func(workflow.Context, *app.Workflow) (*app.GenerateStepsResult, error) {
				return &app.GenerateStepsResult{
					Groups: []*app.WorkflowStepGroup{{
						GroupIdx: 1,
						Status:   app.CompositeStatus{Status: app.StatusPending},
					}},
					Steps: []*app.WorkflowStep{{
						Name:          "generated-step",
						Idx:           100,
						GroupIdx:      1,
						ExecutionType: app.WorkflowStepExecutionTypeSystem,
						Status:        app.CompositeStatus{Status: app.StatusPending},
						QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
					}},
				}, nil
			},
		}
	})

	workflowQueue := e.createTestQueue(ctx, ownerID, ownerType, "install-workflows")
	e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-step-groups")
	e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-steps")
	e.createTestQueue(ctx, ownerID, ownerType, "install-signals")
	e.createTestQueue(ctx, ownerID, ownerType, "install-generate-steps")

	flw := e.createTestWorkflow(ctx, ownerID, ownerType, workflowType, &generateworkflowsteps.Signal{})
	e.enqueueResidentFlow(ctx, workflowQueue.ID, flw, ownerID, ownerType, time.Second)
	e.waitForQueueSignalStatus(ctx, flw.ID, "install_workflows", executeflow.SignalType, app.StatusSuccess)

	flw = e.getWorkflow(ctx, flw.ID)
	require.Equal(e.T(), app.StatusSuccess, flw.Status.Status)
	require.Len(e.T(), flw.Steps, 1)
	require.Equal(e.T(), app.StatusSuccess, flw.Steps[0].Status.Status)
}

func (e *FlowTestSuite) TestResidentEagerGenerationNeverFinishesWithPendingSteps() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	workflowType := app.WorkflowType("test_resident_eager_generated_flow")

	generateworkflowsteps.RegisterGenerators(ownerType, func() map[app.WorkflowType]flow.WorkflowStepGenerator {
		return map[app.WorkflowType]flow.WorkflowStepGenerator{
			workflowType: func(workflow.Context, *app.Workflow) (*app.GenerateStepsResult, error) {
				return &app.GenerateStepsResult{
					Groups: []*app.WorkflowStepGroup{
						{GroupIdx: 1, EagerExecution: true, Status: app.CompositeStatus{Status: app.StatusPending}},
						{GroupIdx: 2, Status: app.CompositeStatus{Status: app.StatusPending}},
					},
					Steps: []*app.WorkflowStep{
						{
							Name:          "eager-step",
							Idx:           100,
							GroupIdx:      1,
							ExecutionType: app.WorkflowStepExecutionTypeSystem,
							Status:        app.CompositeStatus{Status: app.StatusPending},
							QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
						},
						{
							Name:          "non-eager-blocking-step",
							Idx:           200,
							GroupIdx:      2,
							ExecutionType: app.WorkflowStepExecutionTypeSystem,
							Status:        app.CompositeStatus{Status: app.StatusPending},
							QueueSignal:   &signaldb.SignalData{Signal: &CancellableTestSignal{}},
						},
						{
							Name:          "non-eager-apply-step",
							Idx:           300,
							GroupIdx:      2,
							ExecutionType: app.WorkflowStepExecutionTypeSystem,
							Status:        app.CompositeStatus{Status: app.StatusPending},
							QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
						},
					},
				}, nil
			},
		}
	})

	workflowQueue := e.createTestQueue(ctx, ownerID, ownerType, "install-workflows")
	e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-step-groups")
	e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-steps")
	e.createTestQueue(ctx, ownerID, ownerType, "install-signals")
	e.createTestQueue(ctx, ownerID, ownerType, "install-generate-steps")

	flw := e.createTestWorkflow(ctx, ownerID, ownerType, workflowType, &generateworkflowsteps.Signal{})
	e.enqueueResidentFlow(ctx, workflowQueue.ID, flw, ownerID, ownerType, time.Second)
	e.waitForStepInProgress(ctx, flw.ID, "non-eager-blocking-step")

	observedWorkflow := e.getWorkflow(ctx, flw.ID)
	observedQueueSignal := e.getLatestQueueSignal(ctx, flw.ID, "install_workflows", executeflow.SignalType)
	observedSteps := e.getStepsByWorkflow(ctx, flw.ID)

	_, err := e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{InstallWorkflowID: flw.ID})
	require.NoError(e.T(), err)
	e.waitForWorkflowTerminal(ctx, flw.ID)

	require.NotEqual(e.T(), app.StatusSuccess, observedWorkflow.Status.Status)
	require.True(e.T(), observedWorkflow.FinishedAt.IsZero())
	require.NotEqual(e.T(), app.StatusSuccess, observedQueueSignal.Status.Status)
	for _, step := range observedSteps {
		if step.Name == "non-eager-apply-step" {
			require.False(e.T(), isTerminal(step.Status.Status))
		}
	}
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
