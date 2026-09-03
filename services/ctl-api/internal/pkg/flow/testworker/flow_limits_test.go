package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// Lifeline tests for PR #2320: cancelled flows write a terminal cancelled
// status, approval waits and abandoned parks expire at callback.MaxWaitCeiling
// (shrunk to 15s in SetupSuite), and every stopped flow drains its Temporal
// workflows. Expected to fail until this branch includes those fixes.

func parkedFailingStep(name string) app.WorkflowStep {
	return app.WorkflowStep{
		Name:          name,
		Idx:           100,
		GroupIdx:      1,
		ExecutionType: app.WorkflowStepExecutionTypeSystem,
		Retryable:     true,
		QueueSignal:   &signaldb.SignalData{Signal: &ManualRetrySignal{}},
	}
}

// Cancelling a workflow must leave it in a terminal cancelled status with a
// finished-at time — not wedged at failed-pending-retry / running.
func (e *FlowTestSuite) TestCancelledWorkflowHasCancelledStatusAndFinishedAt() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{parkedFailingStep("cancel-writes-status")}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	_, err := e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{InstallWorkflowID: flw.ID})
	require.NoError(e.T(), err)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusCancelled)
	got := e.getWorkflow(ctx, flw.ID)
	require.Equal(e.T(), "workflow cancelled", got.Status.StatusHumanDescription)
	e.waitForWorkflowFinished(ctx, flw.ID)
	e.assertTemporalDrained(ctx, flw.ID)
}

// An approval response arriving in time lets the workflow continue, complete,
// and release its Temporal workflows.
func (e *FlowTestSuite) TestApprovalReceivedWorkflowCompletes() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("approve-completes", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
		{
			Name:          "after-approval",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	approval := e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)
	e.respondApproval(ctx, flw, &steps[0], approval.ID, app.WorkflowStepApprovalResponseTypeApprove)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	require.False(e.T(), e.getWorkflow(ctx, flw.ID).FinishedAt.IsZero(),
		"a completed workflow must be finished")
	e.assertTemporalDrained(ctx, flw.ID)
}

// With no approval response, the wait expires at MaxWaitCeiling: the step is
// marked approval-expired with a stop directive, no retry clone is created,
// and the workflow finishes and drains instead of living forever.
func (e *FlowTestSuite) TestApprovalExpiresStopsWorkflow() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("approval-expires", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)

	require.Eventually(e.T(), func() bool {
		return e.getStep(ctx, steps[0].ID).Status.Status == app.WorkflowStepApprovalStatusApprovalExpired
	}, ceilingWait, pollInterval, "step did not expire its approval wait")

	step := e.getStep(ctx, steps[0].ID)
	require.Equal(e.T(), "no approval received", step.Status.StatusHumanDescription)
	require.Equal(e.T(), directive.StepStop, directive.Step(step.ResultDirective))
	require.Len(e.T(), e.getStepsByWorkflow(ctx, flw.ID), 1,
		"an expired approval must not spawn retry clones")

	require.Eventually(e.T(), func() bool {
		return !e.getWorkflow(ctx, flw.ID).FinishedAt.IsZero()
	}, ceilingWait, pollInterval, "workflow with expired approval must finish")
	e.assertTemporalDrained(ctx, flw.ID)
}

// Denying the approval ends the workflow: terminal status, finished-at set,
// Temporal workflows closed.
func (e *FlowTestSuite) TestApprovalDeniedStopsWorkflow() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("deny-stops", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	approval := e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)
	e.respondApproval(ctx, flw, &steps[0], approval.ID, app.WorkflowStepApprovalResponseTypeDeny)

	e.waitForStepStatus(ctx, steps[0].ID, app.WorkflowStepApprovalStatusApprovalDenied)
	require.Eventually(e.T(), func() bool {
		return !e.getWorkflow(ctx, flw.ID).FinishedAt.IsZero()
	}, ceilingWait, pollInterval, "denied workflow must finish")
	e.assertTemporalDrained(ctx, flw.ID)
}

// A parked step that never receives a retry or skip is abandoned at
// MaxWaitCeiling: marked errored with the abandoned reason and stop directive,
// and the workflow finishes and drains.
func (e *FlowTestSuite) TestParkedRetryExpiresStopsWorkflow() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{parkedFailingStep("park-expires")}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	require.Eventually(e.T(), func() bool {
		step := e.getStep(ctx, steps[0].ID)
		return step.Status.Status == app.StatusError &&
			step.Status.StatusHumanDescription == "step abandoned: no retry or skip received" &&
			directive.Step(step.ResultDirective) == directive.StepStop
	}, ceilingWait, pollInterval, "parked step was not abandoned at the wait ceiling")

	step := e.getStep(ctx, steps[0].ID)
	require.Equal(e.T(), true, step.Status.Metadata["abandoned"])
	require.Equal(e.T(), directive.StepStop, directive.Step(step.ResultDirective))

	require.Eventually(e.T(), func() bool {
		return !e.getWorkflow(ctx, flw.ID).FinishedAt.IsZero()
	}, ceilingWait, pollInterval, "workflow with abandoned step must finish")
	e.assertTemporalDrained(ctx, flw.ID)
}
