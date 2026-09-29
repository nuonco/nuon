package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

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

func (e *FlowTestSuite) TestLegacyApprovalExpiresStopsWorkflow() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("approval-expires", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.seedApproval(ctx, &steps[0])

	e.enqueueLegacyFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)

	require.Eventually(e.T(), func() bool {
		step := e.getStep(ctx, steps[0].ID)
		return step.Status.Status == app.WorkflowStepApprovalStatusApprovalExpired &&
			step.Status.StatusHumanDescription == "no approval received" &&
			directive.Step(step.ResultDirective) == directive.StepStop
	}, ceilingWait, pollInterval, "step did not expire its approval wait")

	require.Len(e.T(), e.getStepsByWorkflow(ctx, flw.ID), 1,
		"an expired approval must not spawn retry clones")

	require.Eventually(e.T(), func() bool {
		var wf app.Workflow
		if err := e.service.DB.WithContext(ctx).First(&wf, "id = ?", flw.ID).Error; err != nil {
			return false
		}
		return !wf.FinishedAt.IsZero()
	}, ceilingWait, pollInterval, "workflow with expired approval must finish")
	e.assertTemporalDrained(ctx, flw.ID)
}

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
		var wf app.Workflow
		if err := e.service.DB.WithContext(ctx).First(&wf, "id = ?", flw.ID).Error; err != nil {
			return false
		}
		return !wf.FinishedAt.IsZero()
	}, ceilingWait, pollInterval, "denied workflow must finish")
	e.assertTemporalDrained(ctx, flw.ID)
}

func (e *FlowTestSuite) TestParkedRetryExpiresStopsWorkflow() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{parkedFailingStep("park-expires")}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForResidentAwaitRetry(ctx, flw, "park-expires")
	e.waitForQueueSignalStatus(ctx, flw.ID, "install_workflows", executeflow.SignalType, app.StatusSuccess)
	e.assertTemporalDrained(ctx, flw.ID)

	step := e.getStep(ctx, steps[0].ID)
	require.Equal(e.T(), app.StatusError, step.Status.Status)
	require.Equal(e.T(), directive.StepAwaitRetry, directive.Step(step.ResultDirective))
	require.Equal(e.T(), app.StatusFailedPendingRetry, e.getWorkflow(ctx, flw.ID).Status.Status)
}
