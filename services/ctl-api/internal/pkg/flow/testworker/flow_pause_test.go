package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (e *FlowTestSuite) TestPauseAndUnpause() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "g1-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "g2-step", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "g3-step", Idx: 300, GroupIdx: 3, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	err := e.service.FlowClient.PauseWorkflow(ctx, &flowclient.PauseWorkflowRequest{
		InstallWorkflowID: flw.ID,
	})
	_ = err

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	e.waitForStepStatus(ctx, e.getStepsByWorkflow(ctx, flw.ID)[0].ID, app.StatusSuccess)

	err = e.service.FlowClient.PauseWorkflow(ctx, &flowclient.PauseWorkflowRequest{
		InstallWorkflowID: flw.ID,
	})
	require.Nil(e.T(), err)

	err = e.service.FlowClient.UnpauseWorkflow(ctx, &flowclient.UnpauseWorkflowRequest{
		InstallWorkflowID: flw.ID,
	})
	require.Nil(e.T(), err)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		require.Equal(e.T(), app.StatusSuccess, step.Status.Status,
			"step %s should be success after unpause", step.Name)
	}
	e.assertTemporalDrained(ctx, flw.ID)
}
