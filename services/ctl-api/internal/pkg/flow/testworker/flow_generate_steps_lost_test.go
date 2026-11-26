package testworker

import (
	"sync/atomic"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateworkflowsteps"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (e *FlowTestSuite) TestGenerateStepsReenqueuesWhenHandlerRunFails() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	workflowType := app.WorkflowType("test_generate_steps_handler_lost_flow")

	var calls atomic.Int32
	registerTestGenerator(ownerType, workflowType, func(workflow.Context, *app.Workflow) (*app.GenerateStepsResult, error) {
		if calls.Add(1) == 1 {
			time.Sleep(3 * time.Second)
		}
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
					Name:          "later-step",
					Idx:           200,
					GroupIdx:      2,
					ExecutionType: app.WorkflowStepExecutionTypeSystem,
					Status:        app.CompositeStatus{Status: app.StatusPending},
					QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
				},
			},
		}, nil
	})

	workflowQueue := e.createTestQueue(ctx, ownerID, ownerType, "install-workflows")
	e.createTestQueue(ctx, ownerID, ownerType, queuenames.InstallWorkflowStepGroupsQueueName)
	e.createTestQueue(ctx, ownerID, ownerType, queuenames.InstallWorkflowStepsQueueName)
	e.createTestQueue(ctx, ownerID, ownerType, queuenames.InstallSignalsQueueName)
	e.createTestQueue(ctx, ownerID, ownerType, queuenames.InstallGenerateStepsQueueName)

	flw := e.createTestWorkflow(ctx, ownerID, ownerType, workflowType, &generateworkflowsteps.Signal{})
	e.enqueueFlow(ctx, workflowQueue.ID, flw, ownerID, ownerType)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	require.GreaterOrEqual(e.T(), calls.Load(), int32(2))

	var generateSignals []app.QueueSignal
	res := e.service.DB.WithContext(ctx).
		Where(app.QueueSignal{OwnerID: flw.ID, OwnerType: "install_workflows", Type: generateworkflowsteps.SignalType}).
		Order("created_at asc").
		Find(&generateSignals)
	require.NoError(e.T(), res.Error)
	require.Len(e.T(), generateSignals, 2)
	require.NotEqual(e.T(), app.StatusSuccess, generateSignals[0].Status.Status)
	require.Equal(e.T(), app.StatusSuccess, generateSignals[1].Status.Status)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	require.Len(e.T(), steps, 2)
	for _, step := range steps {
		require.Equal(e.T(), app.StatusSuccess, step.Status.Status, step.Name)
	}
	e.assertTemporalDrained(ctx, flw.ID)
}
