package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (e *FlowTestSuite) TestCancelStepCallsInnerCancel() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "cancellable-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &CancellableTestSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	stepID := e.waitForStepInProgress(ctx, flw.ID, "cancellable-step")
	e.waitForQueueSignalStatus(ctx, stepID, "install_workflow_steps", CancellableTestSignalType, app.StatusInProgress)

	_, err := e.service.FlowClient.CancelStep(ctx, &flowclient.CancelStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            stepID,
	})
	require.Nil(e.T(), err)

	e.waitForWorkflowTerminal(ctx, flw.ID)

	step := e.getStep(ctx, stepID)
	require.Equal(e.T(), CancelMarker, step.ResultDirective,
		"inner signal Cancel() should have written the cancel marker to ResultDirective")
	e.assertTemporalDrained(ctx, flw.ID)
}

func (e *FlowTestSuite) TestCancelWorkflowPropagatesDown() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "cancellable-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &CancellableTestSignal{}}},
		{Name: "after-cancel", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	e.waitForStepInProgress(ctx, flw.ID, "cancellable-step")

	_, err := e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{
		InstallWorkflowID: flw.ID,
	})
	require.Nil(e.T(), err)

	e.waitForWorkflowTerminal(ctx, flw.ID)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		switch step.Name {
		case "cancellable-step":
			e.waitForStepStatus(ctx, step.ID, app.StatusCancelled)
		case "after-cancel":
			require.NotEqual(e.T(), app.StatusSuccess, step.Status.Status,
				"step after cancel should not have executed")
		}
	}
	e.assertTemporalDrained(ctx, flw.ID)
}

func (e *FlowTestSuite) TestCancelGroupPropagatesDown() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "cancellable-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &CancellableTestSignal{}}},
		{Name: "g2-step", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	stepID := e.waitForStepInProgress(ctx, flw.ID, "cancellable-step")

	_, err := e.service.FlowClient.CancelGroup(ctx, &flowclient.CancelGroupRequest{
		InstallWorkflowID: flw.ID,
		StepID:            stepID,
	})
	require.Nil(e.T(), err)

	e.waitForWorkflowTerminal(ctx, flw.ID)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		if step.Name == "g2-step" {
			require.NotEqual(e.T(), app.StatusSuccess, step.Status.Status,
				"group 2 should not have executed after group 1 cancellation")
		}
	}
	e.assertTemporalDrained(ctx, flw.ID)
}
