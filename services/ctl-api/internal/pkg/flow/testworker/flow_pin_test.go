package testworker

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestPin* tests pin DESIRED behavior main lacks and are EXPECTED TO FAIL; exclude with -skip 'TestSuite/TestPin'.

// pinTimeout keeps intentionally-failing waits cheap.
const pinTimeout = 15 * time.Second

func (e *FlowTestSuite) pinWaitWorkflowStatus(ctx context.Context, workflowID string, expected app.Status) {
	require.Eventually(e.T(), func() bool {
		return e.getWorkflow(ctx, workflowID).Status.Status == expected
	}, pinTimeout, pollInterval, "workflow %s did not reach status %s", workflowID, expected)
}

func (e *FlowTestSuite) pinWaitStepStatus(ctx context.Context, stepID string, expected app.Status) {
	require.Eventually(e.T(), func() bool {
		return e.getStep(ctx, stepID).Status.Status == expected
	}, pinTimeout, pollInterval, "step %s did not reach status %s", stepID, expected)
}

// PIN: deny currently yields generic error "workflow stopped"; desired is a distinct plan-rejected terminal status.
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

// PIN: a parked workflow shows failed-pending-retry; desired is errored whenever a step failed and nothing is progressing.
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

// PIN: cancel-step writes no terminal workflow status on main, wedging the flow at failed-pending-retry forever; desired is cancelled+finished.
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

// PIN: policy evaluation failures are only logged and the step parks as if policies passed; desired is a failed step.
func (e *FlowTestSuite) TestPinPolicyEvaluationFailureFailsStep() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	// Dangling component/build refs make policy-context resolution fail.
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
