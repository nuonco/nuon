package testworker

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// A read-only poll against a completed resident flow starts a fresh Handler
// run via update-with-start, but must never re-drive the conductor: no new
// WorkflowRun rows, no status rewrites.
func (e *FlowTestSuite) TestReadOnlyPollDoesNotRewarmTerminalFlow() {
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
	e.waitForQueueSignalStatus(ctx, flw.ID, "install_workflows", executeflow.SignalType, app.StatusSuccess)
	runsBefore := len(e.getWorkflowRuns(ctx, flw.ID))

	resp, err := e.service.FlowClient.PollNextStep(ctx, &flowclient.PollNextStepRequest{
		InstallWorkflowID: flw.ID,
	})
	require.NoError(e.T(), err)
	require.Empty(e.T(), resp.StepID, "completed flow should report no next step")

	require.Never(e.T(), func() bool {
		return len(e.getWorkflowRuns(ctx, flw.ID)) != runsBefore
	}, 5*time.Second, pollInterval, "read-only poll must not create new workflow runs")

	flowSignal := e.getLatestQueueSignal(ctx, flw.ID, "install_workflows", executeflow.SignalType)
	require.Equal(e.T(), app.StatusSuccess, flowSignal.Status.Status)
	require.Equal(e.T(), app.StatusSuccess, e.getWorkflow(ctx, flw.ID).Status.Status)
}
