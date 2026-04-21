package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// TestStepFailurePreventsNextGroup verifies that when a step fails (without
// auto-retry), the group errors and the next group is never reached.
func (e *FlowTestSuite) TestStepFailurePreventsNextGroup() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "g1-fail", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &FailSignal{Reason: "intentional"}}},
		{Name: "g2-success", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		switch step.Name {
		case "g1-fail":
			require.Equal(e.T(), app.StatusError, step.Status.Status,
				"failing step should have error status")
		case "g2-success":
			require.NotEqual(e.T(), app.StatusSuccess, step.Status.Status,
				"group 2 step should not have executed")
		}
	}
}

// TestStepFailureInGroupDoesNotSkipGroup2 verifies that when a step in group 1
// fails, the second step in the same group is not executed either.
func (e *FlowTestSuite) TestStepFailureInGroupDoesNotSkipGroup2() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "g1-success", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "g1-fail", Idx: 200, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &FailSignal{Reason: "second step fails"}}},
		{Name: "g1-after-fail", Idx: 300, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	steps := e.getStepsByWorkflow(ctx, flw.ID)
	for _, step := range steps {
		switch step.Name {
		case "g1-success":
			require.Equal(e.T(), app.StatusSuccess, step.Status.Status)
		case "g1-fail":
			require.Equal(e.T(), app.StatusError, step.Status.Status)
		case "g1-after-fail":
			// Should not have been executed — still pending or not-attempted
			require.NotEqual(e.T(), app.StatusSuccess, step.Status.Status,
				"step after failure should not have executed")
		}
	}
}
