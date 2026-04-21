package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestCancelStepCallsInnerCancel verifies that cancel-step propagates through
// all three tiers and invokes the inner signal's Cancel() method.
// The CancellableTestSignal writes a marker to ResultDirective in its Cancel()
// method — we check for that marker to prove it ran.
func (e *FlowTestSuite) TestCancelStepCallsInnerCancel() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "cancellable-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &CancellableTestSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	// Wait for the step to be in-progress (inner signal is blocking)
	stepID := e.waitForStepInProgress(ctx, flw.ID, "cancellable-step")

	// Cancel the step via the flow client
	_, err := e.service.FlowClient.CancelStep(ctx, &flowclient.CancelStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            stepID,
	})
	require.Nil(e.T(), err)

	// Wait for the workflow to reach a terminal state
	e.waitForWorkflowTerminal(ctx, flw.ID)

	// Verify the inner signal's Cancel() was called by checking the marker.
	// The CancellableTestSignal writes CancelMarker to ResultDirective in Cancel().
	step := e.getStep(ctx, stepID)
	require.Equal(e.T(), CancelMarker, step.ResultDirective,
		"inner signal Cancel() should have written the cancel marker to ResultDirective")
}

// TestCancelWorkflowPropagatesDown verifies that cancel-workflow stops the
// workflow and cancels in-flight steps.
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

	// Wait for the first step to be in-progress
	e.waitForStepInProgress(ctx, flw.ID, "cancellable-step")

	// Cancel the entire workflow
	_, err := e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{
		InstallWorkflowID: flw.ID,
	})
	require.Nil(e.T(), err)

	// Wait for the workflow to reach a terminal state
	e.waitForWorkflowTerminal(ctx, flw.ID)

	// Verify step statuses
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		switch step.Name {
		case "cancellable-step":
			require.Equal(e.T(), app.StatusCancelled, step.Status.Status,
				"in-flight step should be cancelled")
		case "after-cancel":
			require.NotEqual(e.T(), app.StatusSuccess, step.Status.Status,
				"step after cancel should not have executed")
		}
	}
}

// TestCancelGroupPropagatesDown verifies that cancel-group cancels all steps
// in the specified group and stops the workflow.
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

	// Wait for the step to be in-progress
	stepID := e.waitForStepInProgress(ctx, flw.ID, "cancellable-step")

	// Cancel the group containing the step
	_, err := e.service.FlowClient.CancelGroup(ctx, &flowclient.CancelGroupRequest{
		InstallWorkflowID: flw.ID,
		StepID:            stepID,
	})
	require.Nil(e.T(), err)

	// Wait for the workflow to reach a terminal state
	e.waitForWorkflowTerminal(ctx, flw.ID)

	// Group 2 should not have executed
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		if step.Name == "g2-step" {
			require.NotEqual(e.T(), app.StatusSuccess, step.Status.Status,
				"group 2 should not have executed after group 1 cancellation")
		}
	}
}
