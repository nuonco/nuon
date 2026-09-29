package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (e *FlowTestSuite) TestRetryGroupClonesEntireGroup() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	planSignal := &SuccessSignal{}
	applySignal := &PlanApplyFailSignal{}
	finalizeSignal := &SuccessSignal{}
	e.phase("seed")

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "g1-plan", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: planSignal}},
		{Name: "g1-apply", Idx: 200, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:   true,
			QueueSignal: &signaldb.SignalData{Signal: applySignal}},
		{Name: "g2-finalize", Idx: 300, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: finalizeSignal}},
	})
	e.phase("fixtures")

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)
	e.phase("enqueue")

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)
	e.phase("db-success")

	steps := e.getStepsByWorkflow(ctx, flw.ID)

	groupRetryIdxs := make(map[int]bool)
	for _, step := range steps {
		if step.GroupIdx == 1 {
			groupRetryIdxs[step.GroupRetryIdx] = true
		}
	}

	require.GreaterOrEqual(e.T(), len(groupRetryIdxs), 2,
		"expected multiple group retry generations, got %d: %v", len(groupRetryIdxs), groupRetryIdxs)

	for _, step := range steps {
		if step.GroupIdx == 1 && step.GroupRetryIdx == 0 {
			if step.Name == "g1-plan" {
				require.Equal(e.T(), app.StatusSuccess, step.Status.Status)
			} else {
				require.Equal(e.T(), app.StatusError, step.Status.Status)
				require.True(e.T(), step.Retried)
			}
		}
	}

	for _, step := range steps {
		if step.GroupIdx == 2 {
			require.NotEqual(e.T(), app.StatusSuccess, step.Status.Status,
				"group 2 step should not have succeeded since group 1 never passed")
		}
	}
	e.phase("assertions")

	e.assertTemporalDrained(ctx, flw.ID)
	e.phase("drain")
}

func (e *FlowTestSuite) TestRetryGroupRetryOfRetryDiscardsAllPreviousGroups() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	planSignal := &SuccessSignal{}
	applySignal := &PlanApplyFailSignal{}
	triggerSignal := &SuccessSignal{}

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "g1-plan", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: planSignal}},
		{Name: "g1-apply", Idx: 200, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:   true,
			QueueSignal: &signaldb.SignalData{Signal: applySignal}},
		{Name: "g1-trigger", Idx: 300, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: triggerSignal}},
		{Name: "g2-finalize", Idx: 400, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)

	steps := e.getStepsByWorkflow(ctx, flw.ID)

	groupRetryIdxs := make(map[int]bool)
	for _, step := range steps {
		if step.GroupIdx == 1 {
			groupRetryIdxs[step.GroupRetryIdx] = true
		}
	}
	require.GreaterOrEqual(e.T(), len(groupRetryIdxs), 3,
		"expected at least 3 group retry generations (original + 2 retries), got %d: %v",
		len(groupRetryIdxs), groupRetryIdxs)

	var allGroups []app.WorkflowStepGroup
	res := e.service.DB.WithContext(ctx).
		Where("workflow_id = ? AND group_idx = ?", flw.ID, 1).
		Find(&allGroups)
	require.Nil(e.T(), res.Error)

	nonDiscardedCount := 0
	for _, g := range allGroups {
		if g.Status.Status != app.StatusDiscarded {
			nonDiscardedCount++
		}
	}
	require.LessOrEqual(e.T(), nonDiscardedCount, 1,
		"expected at most 1 non-discarded group for GroupIdx=1, got %d", nonDiscardedCount)

	triggerCount := 0
	for _, step := range steps {
		if step.GroupIdx == 1 && step.Name == "g1-trigger" {
			triggerCount++
		}
	}
	require.GreaterOrEqual(e.T(), triggerCount, 1,
		"trigger step should be present in at least the original generation")
	e.assertTemporalDrained(ctx, flw.ID)
}
