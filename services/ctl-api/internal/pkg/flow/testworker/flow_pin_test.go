package testworker

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

const pinTimeout = 15 * time.Second

func (e *FlowTestSuite) pinWaitWorkflowStatus(ctx context.Context, workflowID string, expected app.Status) {
	require.Eventually(e.T(), func() bool {
		status, ok := e.workflowStatus(ctx, workflowID)
		return ok && status == expected
	}, pinTimeout, pollInterval, "workflow %s did not reach status %s", workflowID, expected)
}

func (e *FlowTestSuite) pinWaitStepStatus(ctx context.Context, stepID string, expected app.Status) {
	require.Eventually(e.T(), func() bool {
		step := &app.WorkflowStep{}
		if err := e.service.DB.WithContext(ctx).First(step, "id = ?", stepID).Error; err != nil {
			return false
		}
		return step.Status.Status == expected
	}, pinTimeout, pollInterval, "step %s did not reach status %s", stepID, expected)
}

func (e *FlowTestSuite) TestPinDenyMarksWorkflowRejected() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("deny-marks-rejected", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	approval := e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)
	e.respondApproval(ctx, flw, &steps[0], approval.ID, app.WorkflowStepApprovalResponseTypeDeny)

	e.pinWaitWorkflowStatus(ctx, flw.ID, app.Status("plan-rejected"))
	require.False(e.T(), e.getWorkflow(ctx, flw.ID).FinishedAt.IsZero(),
		"a rejected workflow must be finished")
}

func (e *FlowTestSuite) TestPinParkedWorkflowShowsErrored() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		{
			Name:          "parked-shows-errored",
			Idx:           100,
			GroupIdx:      1,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:     true,
			QueueSignal:   &signaldb.SignalData{Signal: &ManualRetrySignal{}},
		},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)

	e.waitForStepStatus(ctx, steps[0].ID, app.StatusError)
	e.pinWaitWorkflowStatus(ctx, flw.ID, app.StatusError)
}

func (e *FlowTestSuite) TestPinCancelStepTerminatesWorkflow() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		{
			Name:          "cancel-terminates",
			Idx:           100,
			GroupIdx:      1,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:     true,
			QueueSignal:   &signaldb.SignalData{Signal: &ManualRetrySignal{}},
		},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	_, err := e.service.FlowClient.CancelStep(ctx, &flowclient.CancelStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            steps[0].ID,
	})
	require.NoError(e.T(), err)
	e.waitForStepStatus(ctx, steps[0].ID, app.StatusCancelled)

	e.pinWaitWorkflowStatus(ctx, flw.ID, app.StatusCancelled)
	require.False(e.T(), e.getWorkflow(ctx, flw.ID).FinishedAt.IsZero(),
		"a workflow whose only live step was cancelled must be finished")
}

func (e *FlowTestSuite) TestPinPolicyEvaluationFailureFailsStep() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	deploy := e.seedDeployTarget(ctx, app.InstallDeployStatusActive)

	steps := []app.WorkflowStep{
		approvalStep("policy-eval-fails", 1, signaldb.SignalData{Signal: &PolicyEvalApprovalSignal{}}),
	}
	steps[0].StepTargetType = string(app.WorkflowStepTargetTypeInstallDeploys)
	steps[0].StepTargetID = deploy.ID
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)

	e.pinWaitStepStatus(ctx, steps[0].ID, app.StatusError)
}
