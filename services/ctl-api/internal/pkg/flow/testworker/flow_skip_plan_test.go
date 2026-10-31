package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (e *FlowTestSuite) TestSkipPolicyDeniedPlanDiscardsApply() {
	e.testSkipPlanDiscardsApply(true, true)
}

func (e *FlowTestSuite) TestSkipAwaitingApprovalPlanDiscardsApply() {
	e.testSkipPlanDiscardsApply(false, true)
}

func (e *FlowTestSuite) TestSkipLegacyAwaitingApprovalPlanDiscardsApply() {
	e.testSkipPlanDiscardsApply(false, false)
}

func (e *FlowTestSuite) TestSkipSupersededPlanPreservesReplacement() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("failed plan", 1, signaldb.SignalData{Signal: &SkipGroupApprovalSignal{}}),
		approvalStep("replacement plan", 1, signaldb.SignalData{Signal: &SkipGroupApprovalSignal{}}),
		{Name: "apply database", Idx: 102, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	}
	steps[0].Retried = true
	steps[0].Status.Status = app.StatusError
	steps[1].Idx = 101
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	approval := e.seedApproval(ctx, &steps[1])
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[1].ID)
	resp, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flw.ID, StepID: steps[0].ID,
	})
	require.NoError(e.T(), err)
	require.False(e.T(), resp.Skippable)
	require.Equal(e.T(), app.StatusError, e.getStep(ctx, steps[0].ID).Status.Status)
	require.Equal(e.T(), app.AwaitingApproval, e.getStep(ctx, steps[1].ID).Status.Status)
	require.Equal(e.T(), app.StatusPending, e.getStep(ctx, steps[2].ID).Status.Status)
	e.respondApproval(ctx, flw, &steps[1], approval.ID, app.WorkflowStepApprovalResponseTypeApprove)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	require.Equal(e.T(), app.StatusSuccess, e.getStep(ctx, steps[2].ID).Status.Status)
}

func (e *FlowTestSuite) testSkipPlanDiscardsApply(policyDenied, resident bool) {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("plan database", 1, signaldb.SignalData{Signal: &SkipGroupApprovalSignal{}}),
		{Name: "apply database", Idx: 101, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "post-deploy action", Idx: 102, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "independent component", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.seedApproval(ctx, &steps[0])
	_, err := e.service.QueueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queueID,
		Signal: &executeflow.Signal{
			WorkflowID:             flw.ID,
			StepGroupQueueName:     queuenames.InstallWorkflowStepGroupsQueueName,
			StepQueueName:          queuenames.InstallWorkflowStepsQueueName,
			StepTargetQueueName:    queuenames.InstallSignalsQueueName,
			GenerateStepsQueueName: queuenames.InstallGenerateStepsQueueName,
			OwnerID:                ownerID,
			OwnerType:              ownerType,
			Resident:               resident,
			ResidentIdleTimeout:    testResidentIdleTimeout,
		},
		OwnerID: flw.ID, OwnerType: "install_workflows",
	})
	require.NoError(e.T(), err)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)
	if policyDenied {
		e.assertTemporalDrained(ctx, flw.ID)
		status := app.CompositeStatus{Status: app.StatusError, Metadata: map[string]any{"check": "policy"}}
		require.NoError(e.T(), e.service.DB.WithContext(ctx).Model(&steps[0]).Updates(map[string]any{
			"status": status, "result_directive": "",
		}).Error)
		require.NoError(e.T(), e.service.DB.WithContext(ctx).Model(&app.WorkflowStepGroup{}).
			Where(app.WorkflowStepGroup{ID: steps[0].WorkflowStepGroupID}).
			Updates(map[string]any{"status": status, "result_directive": ""}).Error)
		require.NoError(e.T(), e.service.DB.WithContext(ctx).Model(flw).Update("status", status).Error)
	}
	for _, step := range steps[1:] {
		require.Equal(e.T(), app.StatusPending, e.getStep(ctx, step.ID).Status.Status)
	}
	resp, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flw.ID, StepID: steps[0].ID,
	})
	require.NoError(e.T(), err)
	require.True(e.T(), resp.Skippable)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	require.Equal(e.T(), app.StatusUserSkipped, e.getStep(ctx, steps[0].ID).Status.Status)
	for _, step := range steps[1:3] {
		actual := e.getStep(ctx, step.ID)
		require.Equal(e.T(), app.StatusDiscarded, actual.Status.Status, step.Name)
		require.True(e.T(), actual.StartedAt.IsZero(), "%s must never start", step.Name)
		var dispatched int64
		require.NoError(e.T(), e.service.DB.WithContext(ctx).Model(&app.QueueSignal{}).
			Where(app.QueueSignal{OwnerID: step.ID}).Count(&dispatched).Error)
		require.Zero(e.T(), dispatched, "%s must never be dispatched", step.Name)
	}
	require.Equal(e.T(), app.StatusSuccess, e.getStep(ctx, steps[3].ID).Status.Status)
}
