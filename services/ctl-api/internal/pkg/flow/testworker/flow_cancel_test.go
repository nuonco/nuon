package testworker

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestCancelStepCallsInnerCancel verifies that cancel-step propagates through
// all three tiers and invokes the inner signal's Cancel() method.
func (e *FlowTestSuite) TestCancelStepCallsInnerCancel() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "cancellable-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &CancellableTestSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	// Wait for the inner signal itself to be executing, not just the step: a
	// cancel that lands before the inner handler accepts updates falls back to
	// a direct DB status write and never invokes Cancel().
	stepID := e.waitForStepInProgress(ctx, flw.ID, "cancellable-step")
	require.Eventually(e.T(), func() bool {
		var qs app.QueueSignal
		err := e.service.DB.WithContext(ctx).
			Where(app.QueueSignal{OwnerID: stepID, OwnerType: "install_workflow_steps", Type: CancellableTestSignalType}).
			First(&qs).Error
		_, started := qs.Status.Metadata["execute_started_at"]
		return err == nil && qs.Status.Status == app.StatusInProgress && started
	}, pollTimeout, pollInterval, "inner signal should be executing before cancellation")
	innerSignal := e.getLatestQueueSignal(ctx, stepID, "install_workflow_steps", CancellableTestSignalType)

	// Cancel the step via the flow client
	_, err := e.service.FlowClient.CancelStep(ctx, &flowclient.CancelStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            stepID,
	})
	require.Nil(e.T(), err)

	require.Eventually(e.T(), func() bool {
		var qs app.QueueSignal
		err := e.service.DB.WithContext(ctx).
			Where(app.QueueSignal{ID: innerSignal.ID}).
			First(&qs).Error
		return err == nil && qs.Status.Status == app.StatusCancelled &&
			qs.Status.StatusHumanDescription == "cancel-callback-invoked"
	}, pollTimeout, pollInterval, "inner signal Cancel() should complete successfully")
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusCancelled)
	e.waitForStepStatus(ctx, stepID, app.StatusCancelled)
	e.assertTemporalDrained(ctx, flw.ID)
	require.Equal(e.T(), "stop", e.getStep(ctx, stepID).ResultDirective)
}

func (e *FlowTestSuite) TestCancelStepRetryWithoutTargetCancelsSupersededTarget() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	flowID, deployID, clone := e.parkRetryWithoutTarget(ctx, false)
	require.Equal(e.T(), 1, clone.RetryIndex)

	e.cancelWorkflow(ctx, flowID)
	e.waitForWorkflowStatus(ctx, flowID, app.StatusCancelled)
	e.waitForDeployStatus(ctx, deployID, app.InstallDeployStatusCancelled)
}

func (e *FlowTestSuite) TestCancelGroupRetryWithoutTargetCancelsSupersededTarget() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	flowID, deployID, clone := e.parkRetryWithoutTarget(ctx, true)
	require.Equal(e.T(), 0, clone.RetryIndex)
	require.Equal(e.T(), 1, clone.GroupRetryIdx)

	e.cancelWorkflow(ctx, flowID)
	e.waitForWorkflowStatus(ctx, flowID, app.StatusCancelled)
	e.waitForDeployStatus(ctx, deployID, app.InstallDeployStatusCancelled)
}

func (e *FlowTestSuite) TestSkipRetryWithoutTargetKeepsSupersededTarget() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	flowID, deployID, clone := e.parkRetryWithoutTarget(ctx, false)

	_, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flowID,
		StepID:            clone.ID,
	})
	require.NoError(e.T(), err)
	e.waitForStepStatus(ctx, clone.ID, app.StatusUserSkipped)
	e.waitForWorkflowTerminal(ctx, flowID)
	require.Equal(e.T(), app.InstallDeployStatusRetried, e.getDeploy(ctx, deployID).Status)
}

func (e *FlowTestSuite) parkRetryWithoutTarget(ctx context.Context, group bool) (string, string, *app.WorkflowStep) {
	ownerID, ownerType := newTestOwner()
	deploy := e.seedDeployTarget(ctx, app.InstallDeployStatusError)
	stepID := domains.NewWorkflowStepID()
	var sig signal.Signal = &RetryWithoutTargetSignal{TargetStepID: stepID}
	if group {
		sig = &GroupRetryWithoutTargetSignal{RetryWithoutTargetSignal{TargetStepID: stepID}}
	}
	steps := []app.WorkflowStep{
		failingTargetedStep("retry-without-target", signaldb.SignalData{Signal: sig}, deploy),
	}
	steps[0].ID = stepID
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	_, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            steps[0].ID,
	})
	require.NoError(e.T(), err)

	var cloneID string
	require.Eventually(e.T(), func() bool {
		all, err := e.tryStepsByWorkflow(ctx, flw.ID)
		if err != nil {
			return false
		}
		for _, s := range all {
			if s.ID != steps[0].ID && s.Name == "retry-without-target" && s.Status.Status == app.StatusError {
				cloneID = s.ID
				return true
			}
		}
		return false
	}, pollTimeout, pollInterval, "retry clone did not fail")
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	clone := e.getStep(ctx, cloneID)
	require.Empty(e.T(), clone.StepTargetID)
	require.Equal(e.T(), app.InstallDeployStatusRetried, e.getDeploy(ctx, deploy.ID).Status)
	return flw.ID, deploy.ID, clone
}

func (e *FlowTestSuite) waitForDeployStatus(ctx context.Context, deployID string, want app.InstallDeployStatus) {
	require.Eventually(e.T(), func() bool {
		var got app.InstallDeploy
		err := e.service.DB.WithContext(ctx).First(&got, "id = ?", deployID).Error
		return err == nil && got.Status == want
	}, pollTimeout, pollInterval, "deploy did not reach status %s", want)
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
			e.waitForStepStatus(ctx, step.ID, app.StatusCancelled)
		case "after-cancel":
			require.NotEqual(e.T(), app.StatusSuccess, step.Status.Status,
				"step after cancel should not have executed")
		}
	}
	e.assertTemporalDrained(ctx, flw.ID)
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
	e.assertTemporalDrained(ctx, flw.ID)
}
