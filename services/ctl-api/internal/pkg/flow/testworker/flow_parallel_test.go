package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestParallelGroupExecution verifies that steps within a parallel group
// all execute and complete.
func (e *FlowTestSuite) TestParallelGroupExecution() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "parallel-1", Idx: 100, GroupIdx: 1, GroupParallel: true, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "parallel-2", Idx: 200, GroupIdx: 1, GroupParallel: true, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "parallel-3", Idx: 300, GroupIdx: 1, GroupParallel: true, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		require.Equal(e.T(), app.StatusSuccess, step.Status.Status,
			"parallel step %s should be success", step.Name)
	}
}

// TestMixedParallelAndSequentialGroups verifies a workflow with both a
// sequential group and a parallel group executes correctly.
func (e *FlowTestSuite) TestMixedParallelAndSequentialGroups() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		// Group 1: sequential (plan + apply pattern)
		{Name: "g1-plan", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "g1-apply", Idx: 200, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		// Group 2: parallel (deploy 3 components concurrently)
		{Name: "g2-deploy-a", Idx: 300, GroupIdx: 2, GroupParallel: true, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "g2-deploy-b", Idx: 400, GroupIdx: 2, GroupParallel: true, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "g2-deploy-c", Idx: 500, GroupIdx: 2, GroupParallel: true, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		// Group 3: sequential (finalize)
		{Name: "g3-finalize", Idx: 600, GroupIdx: 3, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	require.Len(e.T(), steps, 6)
	for _, step := range steps {
		require.Equal(e.T(), app.StatusSuccess, step.Status.Status,
			"step %s should be success", step.Name)
	}
}
