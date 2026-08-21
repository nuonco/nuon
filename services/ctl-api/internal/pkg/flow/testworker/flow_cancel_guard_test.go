package testworker

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// Cancelled workflows are terminal: each guard fires an action at one and asserts nothing revives it.

func (e *FlowTestSuite) setupCancelledParkedFlow(ctx context.Context) (*app.Workflow, string) {
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		{
			Name:          "parked-then-cancelled",
			Idx:           100,
			GroupIdx:      1,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:     true,
			Skippable:     true,
			QueueSignal:   &signaldb.SignalData{Signal: &ManualRetrySignal{}},
		},
		{
			Name:          "never-after-cancel",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	_, err := e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{InstallWorkflowID: flw.ID})
	require.NoError(e.T(), err)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusCancelled)
	return flw, steps[0].ID
}

func (e *FlowTestSuite) assertStillCancelled(ctx context.Context, flw *app.Workflow, wantSteps int) {
	require.Never(e.T(), func() bool {
		if e.getWorkflow(ctx, flw.ID).Status.Status != app.StatusCancelled {
			return true
		}
		steps := e.getStepsByWorkflow(ctx, flw.ID)
		if len(steps) != wantSteps {
			return true
		}
		for _, s := range steps {
			switch s.Status.Status {
			case app.StatusInProgress, app.StatusSuccess, app.StatusUserSkipped,
				app.WorkflowStepApprovalStatusApproved:
				return true
			}
		}
		return false
	}, 5*time.Second, pollInterval, "cancelled workflow must not be revived")
}

func (e *FlowTestSuite) TestCancelledWorkflowRejectsRetry() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	flw, stepID := e.setupCancelledParkedFlow(ctx)

	resp, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            stepID,
	})
	if err == nil {
		_ = resp
	}
	e.assertStillCancelled(ctx, flw, 2)
}

func (e *FlowTestSuite) TestCancelledWorkflowRejectsSkip() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	flw, stepID := e.setupCancelledParkedFlow(ctx)

	resp, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            stepID,
	})
	if err == nil {
		require.False(e.T(), resp.Skippable, "skip on a cancelled workflow must be rejected")
	}
	e.assertStillCancelled(ctx, flw, 2)
}

func (e *FlowTestSuite) TestCancelledWorkflowRejectsApprove() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("approval-then-cancelled", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
		{
			Name:          "never-after-cancel",
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

	_, err := e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{InstallWorkflowID: flw.ID})
	require.NoError(e.T(), err)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusCancelled)

	resp := app.WorkflowStepApprovalResponse{
		InstallWorkflowStepApprovalID: approval.ID,
		Type:                          app.WorkflowStepApprovalResponseTypeApprove,
	}
	require.NoError(e.T(), e.service.DB.WithContext(ctx).Create(&resp).Error)
	_ = e.service.FlowClient.ApprovePlan(ctx, &flowclient.ApprovePlanRequest{
		InstallWorkflowID:  flw.ID,
		StepID:             steps[0].ID,
		ApprovalResponseID: resp.ID,
		ResponseType:       app.WorkflowStepApprovalResponseTypeApprove,
	})

	e.assertStillCancelled(ctx, flw, 2)
}
