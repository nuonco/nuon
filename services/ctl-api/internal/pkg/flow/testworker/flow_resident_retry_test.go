package testworker

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstep"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstepgroup"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func residentRetrySteps(skippable bool) []app.WorkflowStep {
	return []app.WorkflowStep{
		{
			Name:          "await-manual-retry",
			Idx:           100,
			GroupIdx:      1,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:     true,
			Skippable:     skippable,
			QueueSignal:   &signaldb.SignalData{Signal: &ManualRetrySignal{}},
		},
		{
			Name:          "after-manual-action",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
}

func (e *FlowTestSuite) waitForResidentAwaitRetry(ctx context.Context, flw *app.Workflow, stepName string) *app.WorkflowStep {
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	var failedStep *app.WorkflowStep
	require.Eventually(e.T(), func() bool {
		steps := e.getStepsByWorkflow(ctx, flw.ID)
		for i := range steps {
			if steps[i].Name == stepName &&
				steps[i].Status.Status == app.StatusError &&
				directive.Step(steps[i].ResultDirective) == directive.StepAwaitRetry {
				failedStep = &steps[i]
				return true
			}
		}
		return false
	}, pollTimeout, pollInterval)

	e.waitForQueueSignalStatus(ctx, failedStep.ID, "install_workflow_steps", executeworkflowstep.SignalType, app.StatusSuccess)
	e.waitForQueueSignalStatus(ctx, failedStep.WorkflowStepGroupID, "workflow_step_groups", executeworkflowstepgroup.SignalType, app.StatusSuccess)
	require.Equal(e.T(), app.StatusFailedPendingRetry, e.getStepGroup(ctx, failedStep.WorkflowStepGroupID).Status.Status)
	require.Eventually(e.T(), func() bool {
		runs := e.getWorkflowRuns(ctx, flw.ID)
		return len(runs) > 0 && runs[len(runs)-1].Status.Status == app.StatusFailedPendingRetry
	}, pollTimeout, pollInterval)

	return failedStep
}

func (e *FlowTestSuite) assertSingleSuccessfulRetry(ctx context.Context, flw *app.Workflow, originalStepID string) {
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	retryCount := 0
	for i := range steps {
		step := &steps[i]
		if step.GroupIdx != 1 || step.ID == originalStepID {
			continue
		}
		retryCount++
		require.Equal(e.T(), 1, step.RetryIndex)
		require.Equal(e.T(), app.StatusSuccess, step.Status.Status)
	}
	require.Equal(e.T(), 1, retryCount, "manual retry should create and execute exactly one clone")

	for _, step := range steps {
		if step.Name == "after-manual-action" {
			require.Equal(e.T(), app.StatusSuccess, step.Status.Status)
		}
	}
}

func (e *FlowTestSuite) TestResidentAwaitRetryUnwindsAndExpires() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, residentRetrySteps(true))

	e.enqueueResidentFlow(ctx, queueID, flw, ownerID, ownerType, 2*time.Second)
	failedStep := e.waitForResidentAwaitRetry(ctx, flw, "await-manual-retry")
	e.waitForQueueSignalStatus(ctx, flw.ID, "install_workflows", executeflow.SignalType, app.StatusSuccess)

	failedStep = e.getStep(ctx, failedStep.ID)
	require.Equal(e.T(), app.StatusError, failedStep.Status.Status)
	require.Equal(e.T(), directive.StepAwaitRetry, directive.Step(failedStep.ResultDirective))
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		if step.Name == "after-manual-action" {
			require.Equal(e.T(), app.StatusPending, step.Status.Status)
		}
	}
}

func (e *FlowTestSuite) TestResidentWarmRetryResumes() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, residentRetrySteps(true))

	e.enqueueResidentFlow(ctx, queueID, flw, ownerID, ownerType, 30*time.Second)
	failedStep := e.waitForResidentAwaitRetry(ctx, flw, "await-manual-retry")
	flowSignal := e.getLatestQueueSignal(ctx, flw.ID, "install_workflows", executeflow.SignalType)
	require.NotEqual(e.T(), app.StatusSuccess, flowSignal.Status.Status)

	_, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStep.ID,
	})
	require.NoError(e.T(), err)
	e.assertSingleSuccessfulRetry(ctx, flw, failedStep.ID)
}

func (e *FlowTestSuite) TestResidentColdRetryRewarms() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, residentRetrySteps(true))

	e.enqueueResidentFlow(ctx, queueID, flw, ownerID, ownerType, 2*time.Second)
	failedStep := e.waitForResidentAwaitRetry(ctx, flw, "await-manual-retry")
	e.waitForQueueSignalStatus(ctx, flw.ID, "install_workflows", executeflow.SignalType, app.StatusSuccess)

	_, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStep.ID,
	})
	require.NoError(e.T(), err)
	e.assertSingleSuccessfulRetry(ctx, flw, failedStep.ID)
}

func (e *FlowTestSuite) TestResidentRetryIsIdempotentAcrossConcurrentUpdates() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, residentRetrySteps(true))

	e.enqueueResidentFlow(ctx, queueID, flw, ownerID, ownerType, 30*time.Second)
	failedStep := e.waitForResidentAwaitRetry(ctx, flw, "await-manual-retry")

	errs := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
				InstallWorkflowID: flw.ID,
				StepID:            failedStep.ID,
			})
			errs <- err
		}()
	}
	for range 2 {
		require.NoError(e.T(), <-errs)
	}

	e.assertSingleSuccessfulRetry(ctx, flw, failedStep.ID)
}

func (e *FlowTestSuite) TestResidentRetriesDistinctStepsInTheSameGroup() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{
			Name:          "first-manual-retry",
			Idx:           100,
			GroupIdx:      1,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:     true,
			QueueSignal:   &signaldb.SignalData{Signal: &ManualRetrySignal{}},
		},
		{
			Name:          "second-manual-retry",
			Idx:           200,
			GroupIdx:      1,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:     true,
			QueueSignal:   &signaldb.SignalData{Signal: &ManualRetrySignal{}},
		},
		{
			Name:          "after-two-retries",
			Idx:           300,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	})

	e.enqueueResidentFlow(ctx, queueID, flw, ownerID, ownerType, 30*time.Second)
	firstStep := e.waitForResidentAwaitRetry(ctx, flw, "first-manual-retry")
	_, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            firstStep.ID,
	})
	require.NoError(e.T(), err)

	secondStep := e.waitForResidentAwaitRetry(ctx, flw, "second-manual-retry")
	_, err = e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            secondStep.ID,
	})
	require.NoError(e.T(), err)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	retries := map[string]int{}
	for _, step := range steps {
		if step.RetryIndex > 0 {
			retries[step.Name]++
			require.Equal(e.T(), app.StatusSuccess, step.Status.Status)
		}
	}
	require.Equal(e.T(), 1, retries["first-manual-retry"])
	require.Equal(e.T(), 1, retries["second-manual-retry"])
}

func (e *FlowTestSuite) TestResidentWarmSkipResumes() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, residentRetrySteps(true))

	e.enqueueResidentFlow(ctx, queueID, flw, ownerID, ownerType, 30*time.Second)
	failedStep := e.waitForResidentAwaitRetry(ctx, flw, "await-manual-retry")
	flowSignal := e.getLatestQueueSignal(ctx, flw.ID, "install_workflows", executeflow.SignalType)
	require.NotEqual(e.T(), app.StatusSuccess, flowSignal.Status.Status)
	resp, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStep.ID,
	})
	require.NoError(e.T(), err)
	require.True(e.T(), resp.Skippable)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	require.Equal(e.T(), app.StatusUserSkipped, e.getStep(ctx, failedStep.ID).Status.Status)
}

func (e *FlowTestSuite) TestResidentColdSkipRewarms() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, residentRetrySteps(true))

	e.enqueueResidentFlow(ctx, queueID, flw, ownerID, ownerType, 2*time.Second)
	failedStep := e.waitForResidentAwaitRetry(ctx, flw, "await-manual-retry")
	e.waitForQueueSignalStatus(ctx, flw.ID, "install_workflows", executeflow.SignalType, app.StatusSuccess)
	resp, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStep.ID,
	})
	require.NoError(e.T(), err)
	require.True(e.T(), resp.Skippable)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	require.Equal(e.T(), app.StatusUserSkipped, e.getStep(ctx, failedStep.ID).Status.Status)
}

func (e *FlowTestSuite) TestResidentWarmCancelStops() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, residentRetrySteps(true))

	e.enqueueResidentFlow(ctx, queueID, flw, ownerID, ownerType, 30*time.Second)
	failedStep := e.waitForResidentAwaitRetry(ctx, flw, "await-manual-retry")
	flowSignal := e.getLatestQueueSignal(ctx, flw.ID, "install_workflows", executeflow.SignalType)
	require.NotEqual(e.T(), app.StatusSuccess, flowSignal.Status.Status)
	_, err := e.service.FlowClient.CancelStep(ctx, &flowclient.CancelStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStep.ID,
	})
	require.NoError(e.T(), err)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusCancelled)
	require.Equal(e.T(), app.StatusCancelled, e.getStep(ctx, failedStep.ID).Status.Status)
	require.Equal(e.T(), app.StatusCancelled, e.getStepGroup(ctx, failedStep.WorkflowStepGroupID).Status.Status)
}

func (e *FlowTestSuite) TestResidentColdCancelRewarms() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, residentRetrySteps(true))

	e.enqueueResidentFlow(ctx, queueID, flw, ownerID, ownerType, 2*time.Second)
	failedStep := e.waitForResidentAwaitRetry(ctx, flw, "await-manual-retry")
	e.waitForQueueSignalStatus(ctx, flw.ID, "install_workflows", executeflow.SignalType, app.StatusSuccess)
	_, err := e.service.FlowClient.CancelStep(ctx, &flowclient.CancelStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            failedStep.ID,
	})
	require.NoError(e.T(), err)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusCancelled)
	require.Equal(e.T(), app.StatusCancelled, e.getStep(ctx, failedStep.ID).Status.Status)
}

func (e *FlowTestSuite) TestResidentColdCancelWorkflowRepairsAwaitRetryStep() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, residentRetrySteps(true))

	e.enqueueResidentFlow(ctx, queueID, flw, ownerID, ownerType, 2*time.Second)
	failedStep := e.waitForResidentAwaitRetry(ctx, flw, "await-manual-retry")
	e.waitForQueueSignalStatus(ctx, flw.ID, "install_workflows", executeflow.SignalType, app.StatusSuccess)
	_, err := e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{
		InstallWorkflowID: flw.ID,
	})
	require.NoError(e.T(), err)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusCancelled)
	e.waitForStepStatus(ctx, failedStep.ID, app.StatusCancelled)
}

func (e *FlowTestSuite) TestLegacyAwaitRetryRemainsParkedInPlace() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, residentRetrySteps(true))

	e.enqueueGroupedLegacyFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	failedStep := &steps[0]
	time.Sleep(2 * time.Second)

	stepSignal := e.getLatestQueueSignal(ctx, failedStep.ID, "install_workflow_steps", executeworkflowstep.SignalType)
	groupSignal := e.getLatestQueueSignal(ctx, failedStep.WorkflowStepGroupID, "workflow_step_groups", executeworkflowstepgroup.SignalType)
	flowSignal := e.getLatestQueueSignal(ctx, flw.ID, "install_workflows", executeflow.SignalType)
	require.NotEqual(e.T(), app.StatusSuccess, stepSignal.Status.Status)
	require.NotEqual(e.T(), app.StatusSuccess, groupSignal.Status.Status)
	require.NotEqual(e.T(), app.StatusSuccess, flowSignal.Status.Status)
}

func (e *FlowTestSuite) TestResidentCleanSuccessFastExits() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	flw, queueID := e.setupGroupedFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{
			Name:          "resident-success",
			Idx:           100,
			GroupIdx:      1,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	})

	e.enqueueResidentFlow(ctx, queueID, flw, ownerID, ownerType, 30*time.Second)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	require.Eventually(e.T(), func() bool {
		queueSignal := e.getLatestQueueSignal(ctx, flw.ID, "install_workflows", executeflow.SignalType)
		return queueSignal.Status.Status == app.StatusSuccess
	}, 10*time.Second, pollInterval, "resident success should not wait for the idle timeout")
}
