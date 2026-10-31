package testworker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
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

func executeResidentSkipTest(ctx workflow.Context, sig executeflow.Signal) error {
	if err := sig.RegisterUpdateHandlers(ctx); err != nil {
		return err
	}
	return sig.Execute(ctx)
}

func (e *FlowTestSuite) startIsolatedSkipFlow(ctx context.Context, flw *app.Workflow, configure func(worker.Worker)) (tclient.Client, tclient.WorkflowRun) {
	ns, err := e.service.TClient.GetNamespaceClient(defaultNamespace)
	require.NoError(e.T(), err)
	taskQueue := "skip-recovery-" + flw.ID
	w := worker.New(ns, taskQueue, worker.Options{DisableRegistrationAliasing: true})
	w.RegisterWorkflow(executeResidentSkipTest)
	for _, acts := range e.service.Activities.AllActivities() {
		w.RegisterActivity(acts)
	}
	configure(w)
	require.NoError(e.T(), w.Start())
	e.T().Cleanup(w.Stop)
	run, err := ns.ExecuteWorkflow(ctx, tclient.StartWorkflowOptions{
		ID: taskQueue, TaskQueue: taskQueue, WorkflowExecutionTimeout: 5 * time.Minute,
	}, executeResidentSkipTest, executeflow.Signal{
		WorkflowID: flw.ID, OwnerID: flw.OwnerID, OwnerType: flw.OwnerType,
		Resident: true, ResidentIdleTimeout: 5 * time.Minute,
		StepGroupQueueName:     queuenames.InstallWorkflowStepGroupsQueueName,
		StepQueueName:          queuenames.InstallWorkflowStepsQueueName,
		StepTargetQueueName:    queuenames.InstallSignalsQueueName,
		GenerateStepsQueueName: queuenames.InstallGenerateStepsQueueName,
	})
	require.NoError(e.T(), err)
	e.T().Cleanup(func() { _ = ns.CancelWorkflow(context.Background(), run.GetID(), run.GetRunID()) })
	return ns, run
}

func (e *FlowTestSuite) awaitSkipFlowIdle(ctx context.Context, ns tclient.Client, run tclient.WorkflowRun, idleCount int) {
	require.Eventually(e.T(), func() bool {
		history := ns.GetWorkflowHistory(ctx, run.GetID(), run.GetRunID(), false, 0)
		count := 0
		for history.HasNext() {
			event, err := history.Next()
			if err != nil {
				return false
			}
			if timer := event.GetTimerStartedEventAttributes(); timer != nil && timer.StartToFireTimeout.AsDuration() == 5*time.Minute {
				count++
			}
		}
		return count >= idleCount
	}, pollTimeout, pollInterval, "resident conductor must enter its idle wait")
}

func (e *FlowTestSuite) TestSkipRecoversCommittedStatusFailure() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("plan database", 1, signaldb.SignalData{Signal: &SkipGroupApprovalSignal{}}),
		{Name: "apply database", Idx: 101, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "independent component", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	}
	flw, _ := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.seedApproval(ctx, &steps[0])
	var injected atomic.Bool
	ns, run := e.startIsolatedSkipFlow(ctx, flw, func(w worker.Worker) {
		for _, acts := range e.service.Activities.AllActivities() {
			if a, ok := acts.(*statusactivities.Activities); ok {
				w.RegisterActivityWithOptions(func(ctx context.Context, req statusactivities.UpdateStatusRequest) error {
					if err := a.PkgStatusUpdateFlowStepStatus(ctx, req); err != nil {
						return err
					}
					if req.ID == steps[0].ID && req.Status.Status == app.StatusUserSkipped && injected.CompareAndSwap(false, true) {
						return temporal.NewNonRetryableApplicationError("committed skip status without acknowledgement", "InjectedFailure", nil)
					}
					return nil
				}, activity.RegisterOptions{Name: "PkgStatusUpdateFlowStepStatus", DisableAlreadyRegisteredCheck: true})
			}
		}
	})
	e.awaitApprovalParked(ctx, flw, steps[0].ID)
	e.awaitSkipFlowIdle(ctx, ns, run, 1)
	skip := func() (*executeflow.SkipStepResponse, error) {
		update, err := ns.UpdateWorkflow(ctx, tclient.UpdateWorkflowOptions{
			WorkflowID: run.GetID(), UpdateName: "skip-step", WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{executeflow.SkipStepRequest{StepID: steps[0].ID}},
		})
		if err != nil {
			return nil, err
		}
		var resp executeflow.SkipStepResponse
		err = update.Get(ctx, &resp)
		return &resp, err
	}
	_, err := skip()
	require.ErrorContains(e.T(), err, "committed skip status without acknowledgement")
	require.Equal(e.T(), app.StatusUserSkipped, e.getStep(ctx, steps[0].ID).Status.Status)
	require.Equal(e.T(), app.StatusPending, e.getStep(ctx, steps[2].ID).Status.Status)
	resp, err := skip()
	require.NoError(e.T(), err)
	require.True(e.T(), resp.Skippable)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	require.Equal(e.T(), app.StatusSuccess, e.getStep(ctx, steps[2].ID).Status.Status)
	require.Equal(e.T(), app.StatusDiscarded, e.getStep(ctx, steps[1].ID).Status.Status)
	var dispatched int64
	require.NoError(e.T(), e.service.DB.WithContext(ctx).Model(&app.QueueSignal{}).
		Where(app.QueueSignal{OwnerID: steps[1].ID}).Count(&dispatched).Error)
	require.Zero(e.T(), dispatched)
}

func (e *FlowTestSuite) TestSkipDuplicatePreservesLaterFailure() {
	for _, sameGroup := range []bool{false, true} {
		name := "later-group"
		if sameGroup {
			name = "same-group"
		}
		e.T().Run(name, func(t *testing.T) {
			e := &FlowTestSuite{t: t, service: e.service, queueCache: map[string]*app.Queue{}}
			ctx := e.service.Seed.EnsureAccount(t.Context(), t)
			ctx = e.service.Seed.EnsureOrg(ctx, t)
			ownerID, ownerType := newTestOwner()
			first := approvalStep("first plan", 1, signaldb.SignalData{Signal: &SkipGroupApprovalSignal{}})
			groupIdx := 2
			if sameGroup {
				first.QueueSignal = &signaldb.SignalData{Signal: &ApprovalInnerSignal{}}
				groupIdx = 1
			}
			steps := []app.WorkflowStep{
				first,
				{Name: "later plan", Idx: 200, GroupIdx: groupIdx, ExecutionType: app.WorkflowStepExecutionTypeSystem,
					Retryable: true, QueueSignal: &signaldb.SignalData{Signal: &ManualRetrySignal{}}},
				{Name: "later apply", Idx: 210, GroupIdx: groupIdx, ExecutionType: app.WorkflowStepExecutionTypeSystem,
					QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
			}
			flw, _ := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
			e.seedApproval(ctx, &steps[0])
			var failSchedule atomic.Bool
			ns, run := e.startIsolatedSkipFlow(ctx, flw, func(w worker.Worker) {
				for _, acts := range e.service.Activities.AllActivities() {
					if a, ok := acts.(*workflowactivities.Activities); ok {
						w.RegisterActivityWithOptions(func(ctx context.Context, workflowID string) ([]app.WorkflowStepGroup, error) {
							if failSchedule.CompareAndSwap(true, false) {
								return nil, temporal.NewNonRetryableApplicationError("scheduling read failed", "InjectedFailure", nil)
							}
							return a.PkgWorkflowsFlowGetFlowStepGroups(ctx, workflowID)
						}, activity.RegisterOptions{Name: "PkgWorkflowsFlowGetFlowStepGroups", DisableAlreadyRegisteredCheck: true})
					}
				}
			})
			e.awaitApprovalParked(ctx, flw, steps[0].ID)
			e.awaitSkipFlowIdle(ctx, ns, run, 1)
			skip := func(expectReadError bool) {
				update, err := ns.UpdateWorkflow(ctx, tclient.UpdateWorkflowOptions{
					WorkflowID: run.GetID(), UpdateName: "skip-step", WaitForStage: tclient.WorkflowUpdateStageCompleted,
					Args: []any{executeflow.SkipStepRequest{StepID: steps[0].ID}},
				})
				require.NoError(t, err)
				var resp executeflow.SkipStepResponse
				err = update.Get(ctx, &resp)
				if expectReadError {
					require.ErrorContains(t, err, "scheduling read failed")
					return
				}
				require.NoError(t, err)
				require.True(t, resp.Skippable)
			}
			skip(false)
			e.waitForResidentAwaitRetry(ctx, flw, "later plan")
			e.awaitSkipFlowIdle(ctx, ns, run, 2)
			failSchedule.Store(true)
			skip(true)
			skip(false)
			require.Equal(t, app.StatusFailedPendingRetry, e.getStepGroup(ctx, steps[1].WorkflowStepGroupID).Status.Status)
			require.Never(t, func() bool {
				var dispatched int64
				require.NoError(t, e.service.DB.WithContext(ctx).Model(&app.QueueSignal{}).
					Where(app.QueueSignal{OwnerID: steps[2].ID}).Count(&dispatched).Error)
				return dispatched > 0
			}, 5*time.Second, pollInterval, "a duplicate skip must not dispatch past a later failure")
			require.Equal(t, app.StatusError, e.getStep(ctx, steps[1].ID).Status.Status)
			require.Equal(t, app.StatusPending, e.getStep(ctx, steps[2].ID).Status.Status)
		})
	}
}

func (e *FlowTestSuite) TestSkipSerializesConcurrentRetry() {
	for _, updateName := range []string{"retry-step", "retry-group"} {
		e.T().Run(updateName, func(t *testing.T) {
			e := &FlowTestSuite{t: t, service: e.service, queueCache: map[string]*app.Queue{}}
			ctx := e.service.Seed.EnsureAccount(t.Context(), t)
			ctx = e.service.Seed.EnsureOrg(ctx, t)
			ownerID, ownerType := newTestOwner()
			steps := []app.WorkflowStep{
				approvalStep("plan database", 1, signaldb.SignalData{Signal: &SkipGroupApprovalSignal{}}),
				{Name: "apply database", Idx: 110, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
					QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
			}
			flw, _ := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
			e.seedApproval(ctx, &steps[0])
			readDone, releaseRead := make(chan struct{}), make(chan struct{})
			var reads atomic.Int32
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseRead) }) }
			defer release()
			ns, run := e.startIsolatedSkipFlow(ctx, flw, func(w worker.Worker) {
				for _, acts := range e.service.Activities.AllActivities() {
					if a, ok := acts.(*workflowactivities.Activities); ok {
						w.RegisterActivityWithOptions(func(ctx context.Context, req workflowactivities.GetFlowStepRequest) (*app.WorkflowStep, error) {
							step, err := a.PkgWorkflowsFlowGetFlowsStep(ctx, req)
							if err == nil && req.FlowStepID == steps[0].ID && reads.Add(1) == 2 {
								close(readDone)
								select {
								case <-releaseRead:
								case <-ctx.Done():
									return nil, ctx.Err()
								}
							}
							return step, err
						}, activity.RegisterOptions{Name: "PkgWorkflowsFlowGetFlowsStep", DisableAlreadyRegisteredCheck: true})
					}
				}
			})
			e.awaitApprovalParked(ctx, flw, steps[0].ID)
			skip, err := ns.UpdateWorkflow(ctx, tclient.UpdateWorkflowOptions{
				WorkflowID: run.GetID(), UpdateName: "skip-step", WaitForStage: tclient.WorkflowUpdateStageAccepted,
				Args: []any{executeflow.SkipStepRequest{StepID: steps[0].ID}},
			})
			require.NoError(t, err)
			select {
			case <-readDone:
			case <-time.After(pollTimeout):
				t.Fatal("skip did not reach its eligibility read")
			}
			retry, err := ns.UpdateWorkflow(ctx, tclient.UpdateWorkflowOptions{
				WorkflowID: run.GetID(), UpdateName: updateName, WaitForStage: tclient.WorkflowUpdateStageAccepted,
				Args: []any{executeflow.RetryStepRequest{StepID: steps[0].ID}},
			})
			require.NoError(t, err)
			var retryResponse executeflow.RetryStepResponse
			retryDone := make(chan error, 1)
			go func() { retryDone <- retry.Get(ctx, &retryResponse) }()
			select {
			case err := <-retryDone:
				require.NoError(t, err)
				t.Fatal("retry completed while skip still held its pre-retry snapshot")
			case <-time.After(10 * time.Second):
			}
			release()
			var skipResponse executeflow.SkipStepResponse
			require.NoError(t, skip.Get(ctx, &skipResponse))
			require.True(t, skipResponse.Skippable)
			require.NoError(t, <-retryDone)
			if updateName == "retry-step" {
				require.False(t, retryResponse.Retryable)
				e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
				require.Len(t, e.getStepsByWorkflow(ctx, flw.ID), 2)
				require.Equal(t, app.StatusDiscarded, e.getStep(ctx, steps[1].ID).Status.Status)
			} else {
				require.True(t, retryResponse.Retryable)
				current := e.getStepsByWorkflow(ctx, flw.ID)
				require.Len(t, current, 4)
				for _, step := range current {
					if step.GroupRetryIdx == 1 {
						require.NotEqual(t, app.StatusDiscarded, step.Status.Status)
						if step.Name == "apply database" {
							require.True(t, step.StartedAt.IsZero())
						}
					}
				}
			}
		})
	}
}
