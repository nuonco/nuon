package testworker

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// seedApproval creates the approval row waitForApprovalResponse re-reads from DB; must exist before responding.
func (e *FlowTestSuite) seedApproval(ctx context.Context, step *app.WorkflowStep) *app.WorkflowStepApproval {
	approval := app.WorkflowStepApproval{
		InstallWorkflowStepID: step.ID,
		OwnerID:               step.ID,
		OwnerType:             "install_workflow_steps",
		Type:                  app.NoopApprovalType,
	}
	res := e.service.DB.WithContext(ctx).Create(&approval)
	require.NoError(e.T(), res.Error)
	return &approval
}

// respondApproval mimics the API layer: retry-plan dispatches via RetryStep, everything else via ApprovePlan.
func (e *FlowTestSuite) respondApproval(ctx context.Context, flw *app.Workflow, step *app.WorkflowStep, approvalID string, typ app.WorkflowStepResponseType) *app.WorkflowStepApprovalResponse {
	resp := app.WorkflowStepApprovalResponse{
		InstallWorkflowStepApprovalID: approvalID,
		Type:                          typ,
	}
	res := e.service.DB.WithContext(ctx).Create(&resp)
	require.NoError(e.T(), res.Error)

	if typ == app.WorkflowStepApprovalResponseTypeRetryPlan {
		_, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
			InstallWorkflowID: flw.ID,
			StepID:            step.ID,
		})
		require.NoError(e.T(), err)
		return &resp
	}

	err := e.service.FlowClient.ApprovePlan(ctx, &flowclient.ApprovePlanRequest{
		InstallWorkflowID:  flw.ID,
		StepID:             step.ID,
		ApprovalResponseID: resp.ID,
		ResponseType:       typ,
	})
	require.NoError(e.T(), err)
	return &resp
}

func (e *FlowTestSuite) awaitApprovalParked(ctx context.Context, flw *app.Workflow, stepID string) {
	e.waitForStepStatus(ctx, stepID, app.AwaitingApproval)
	e.waitForWorkflowStatus(ctx, flw.ID, app.AwaitingApproval)
}

func approvalStep(name string, groupIdx int, inner signaldb.SignalData) app.WorkflowStep {
	return app.WorkflowStep{
		Name:          name,
		Idx:           groupIdx * 100,
		GroupIdx:      groupIdx,
		ExecutionType: app.WorkflowStepExecutionTypeApproval,
		Retryable:     true,
		Skippable:     true,
		QueueSignal:   &inner,
	}
}

// An approval step parks step and workflow in approval-awaiting; later steps stay pending.
func (e *FlowTestSuite) TestApprovalAwaitingStatuses() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("await-approval-step", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
		{
			Name:          "after-approval",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)

	step := e.getStep(ctx, steps[0].ID)
	require.Equal(e.T(), directive.StepAwaitApproval, directive.Step(step.ResultDirective))
	for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
		if s.Name == "after-approval" {
			require.Equal(e.T(), app.StatusPending, s.Status.Status)
		}
	}
}

// Approving completes the step and the workflow runs to completion.
func (e *FlowTestSuite) TestApprovalApproveContinues() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("approve-me", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
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
	e.assertStatusMatrix(ctx, steps[0].ID, statusMatrix{
		Step:     app.StatusSuccess,
		Workflow: app.StatusSuccess,
		Run:      app.StatusSuccess,
	})
	for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
		require.Equal(e.T(), app.StatusSuccess, s.Status.Status, s.Name)
	}
	e.assertTemporalDrained(ctx, flw.ID)
}

// After approval the workflow reports in-progress while the next step runs.
func (e *FlowTestSuite) TestApprovalApproveMarksWorkflowRunning() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("approve-then-run", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
		{
			Name:          "blocking-after-approval",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &CancellableTestSignal{}},
		},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	approval := e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)

	e.respondApproval(ctx, flw, &steps[0], approval.ID, app.WorkflowStepApprovalResponseTypeApprove)

	e.waitForStepInProgress(ctx, flw.ID, "blocking-after-approval")
	require.Equal(e.T(), app.StatusInProgress, e.getWorkflow(ctx, flw.ID).Status.Status,
		"workflow must report in-progress while the post-approval step runs")

	_, err := e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{InstallWorkflowID: flw.ID})
	require.NoError(e.T(), err)
	e.waitForWorkflowTerminal(ctx, flw.ID)
	e.assertTemporalDrained(ctx, flw.ID)
}

// Deny stops everything: siblings discarded, later groups not-attempted, target approval-denied, queue signal errors.
func (e *FlowTestSuite) TestApprovalDenyStopsWorkflow() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	deploy := e.seedDeployTarget(ctx, app.InstallDeployStatusActive)

	steps := []app.WorkflowStep{
		approvalStep("deny-me", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
		{
			Name:          "same-group-sibling",
			Idx:           150,
			GroupIdx:      1,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
		{
			Name:          "later-group",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	steps[0].StepTargetType = string(app.WorkflowStepTargetTypeInstallDeploys)
	steps[0].StepTargetID = deploy.ID
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	approval := e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)

	e.respondApproval(ctx, flw, &steps[0], approval.ID, app.WorkflowStepApprovalResponseTypeDeny)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)
	e.assertStatusMatrix(ctx, steps[0].ID, statusMatrix{
		Step:     app.WorkflowStepApprovalStatusApprovalDenied,
		Workflow: app.StatusError,
		Run:      app.StatusError,
		Target:   app.Status(app.InstallDeployApprovalDenied),
	})
	for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
		switch s.Name {
		case "same-group-sibling":
			require.Equal(e.T(), app.StatusDiscarded, s.Status.Status)
		case "later-group":
			require.Equal(e.T(), app.StatusNotAttempted, s.Status.Status)
		}
	}
	require.Contains(e.T(), e.getWorkflow(ctx, flw.ID).Status.StatusHumanDescription, "workflow stopped")
	e.waitForQueueSignalStatus(ctx, flw.ID, "install_workflows", executeflow.SignalType, app.StatusError)
	e.assertTemporalDrained(ctx, flw.ID)
}

// Deny-skip-current denies the step, skips same-group siblings, and the workflow continues.
func (e *FlowTestSuite) TestApprovalDenySkipCurrent() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	deploy := e.seedDeployTarget(ctx, app.InstallDeployStatusActive)

	steps := []app.WorkflowStep{
		approvalStep("skip-current", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
		{
			Name:          "same-group-sibling",
			Idx:           150,
			GroupIdx:      1,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
		{
			Name:          "later-group",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	steps[0].StepTargetType = string(app.WorkflowStepTargetTypeInstallDeploys)
	steps[0].StepTargetID = deploy.ID
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	approval := e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)

	e.respondApproval(ctx, flw, &steps[0], approval.ID, app.WorkflowStepApprovalResponseTypeSkipCurrent)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	e.assertStatusMatrix(ctx, steps[0].ID, statusMatrix{
		Step:   app.StatusSuccess,
		Target: app.Status(app.InstallDeployApprovalDenied),
	})
	for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
		switch s.Name {
		case "same-group-sibling":
			require.Equal(e.T(), app.StatusUserSkipped, s.Status.Status)
		case "later-group":
			require.Equal(e.T(), app.StatusSuccess, s.Status.Status)
		}
	}
	e.assertTemporalDrained(ctx, flw.ID)
}

// Deny-skip-current on a SkipGroup signal skips the whole group; later groups still run.
func (e *FlowTestSuite) TestApprovalDenySkipGroup() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("skip-group-approval", 1, signaldb.SignalData{Signal: &SkipGroupApprovalSignal{}}),
		{
			Name:          "group-mate",
			Idx:           150,
			GroupIdx:      1,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
		{
			Name:          "later-group",
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

	e.respondApproval(ctx, flw, &steps[0], approval.ID, app.WorkflowStepApprovalResponseTypeSkipCurrent)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
		switch s.Name {
		case "group-mate":
			require.Contains(e.T(),
				[]app.Status{app.StatusUserSkipped, app.StatusDiscarded}, s.Status.Status,
				"group mate must not have executed")
		case "later-group":
			require.Equal(e.T(), app.StatusSuccess, s.Status.Status)
		}
	}
	e.assertTemporalDrained(ctx, flw.ID)
}

// Deny-skip-current-and-dependents skips the step then stops the workflow.
func (e *FlowTestSuite) TestApprovalDenySkipDependents() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	deploy := e.seedDeployTarget(ctx, app.InstallDeployStatusActive)

	steps := []app.WorkflowStep{
		approvalStep("skip-dependents", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
		{
			Name:          "later-group",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	steps[0].StepTargetType = string(app.WorkflowStepTargetTypeInstallDeploys)
	steps[0].StepTargetID = deploy.ID
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	approval := e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)

	e.respondApproval(ctx, flw, &steps[0], approval.ID, app.WorkflowStepApprovalResponseTypeSkipCurrentAndDependents)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)
	e.assertStatusMatrix(ctx, steps[0].ID, statusMatrix{
		Workflow: app.StatusError,
		Run:      app.StatusError,
		Target:   app.Status(app.InstallDeployApprovalDenied),
	})
	for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
		if s.Name == "later-group" {
			require.Equal(e.T(), app.StatusNotAttempted, s.Status.Status)
		}
	}
	require.Contains(e.T(), e.getWorkflow(ctx, flw.ID).Status.StatusHumanDescription, "workflow stopped")
	e.assertTemporalDrained(ctx, flw.ID)
}

// Retry-plan supersedes the parked step with exactly one clone that re-parks at approval.
func (e *FlowTestSuite) TestApprovalRetryPlan() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("retry-my-plan", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
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

	e.respondApproval(ctx, flw, &steps[0], approval.ID, app.WorkflowStepApprovalResponseTypeRetryPlan)

	var clone *app.WorkflowStep
	require.Eventually(e.T(), func() bool {
		for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
			if s.ID != steps[0].ID && s.Name == "retry-my-plan" {
				clone = &s
				return s.Status.Status == app.AwaitingApproval
			}
		}
		return false
	}, pollTimeout, pollInterval, "retry-plan must produce exactly one clone that re-parks at approval")

	original := e.getStep(ctx, steps[0].ID)
	require.True(e.T(), original.Retried, "original step must be superseded")
	require.Equal(e.T(), app.StatusDiscarded, original.Status.Status)

	cloneApproval := e.seedApproval(ctx, clone)
	e.respondApproval(ctx, flw, clone, cloneApproval.ID, app.WorkflowStepApprovalResponseTypeApprove)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	cloneCount := 0
	for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
		if s.ID != steps[0].ID && s.Name == "retry-my-plan" {
			cloneCount++
		}
	}
	require.Equal(e.T(), 1, cloneCount, "retry-plan must create exactly one clone")
	e.assertTemporalDrained(ctx, flw.ID)
}

// Cancel-step during approval cancels step+target; main writes no terminal workflow status (see TestPinCancelStepTerminatesWorkflow).
func (e *FlowTestSuite) TestCancelStepDuringApproval() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	deploy := e.seedDeployTarget(ctx, app.InstallDeployStatusActive)

	steps := []app.WorkflowStep{
		approvalStep("cancel-during-approval", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
		{
			Name:          "later-group",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	steps[0].StepTargetType = string(app.WorkflowStepTargetTypeInstallDeploys)
	steps[0].StepTargetID = deploy.ID
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)

	_, err := e.service.FlowClient.CancelStep(ctx, &flowclient.CancelStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            steps[0].ID,
	})
	require.NoError(e.T(), err)

	e.waitForStepStatus(ctx, steps[0].ID, app.StatusCancelled)
	e.assertStatusMatrix(ctx, steps[0].ID, statusMatrix{
		Step:   app.StatusCancelled,
		Target: app.StatusCancelled,
	})
	require.Never(e.T(), func() bool {
		for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
			if s.Name == "later-group" && s.Status.Status == app.StatusSuccess {
				return true
			}
		}
		return false
	}, 5*time.Second, pollInterval, "cancelled workflow must not run further steps")
}

// Cancel-workflow during approval marks the workflow cancelled with finished_at set.
func (e *FlowTestSuite) TestCancelWorkflowDuringApproval() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("cancel-workflow-approval", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
		{
			Name:          "later-group",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)

	_, err := e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{InstallWorkflowID: flw.ID})
	require.NoError(e.T(), err)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusCancelled)
	e.waitForWorkflowFinished(ctx, flw.ID)
	require.Never(e.T(), func() bool {
		for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
			if s.Name == "later-group" && s.Status.Status == app.StatusSuccess {
				return true
			}
		}
		return false
	}, 5*time.Second, pollInterval, "cancelled workflow must not run further steps")
	e.assertTemporalDrained(ctx, flw.ID)
}
