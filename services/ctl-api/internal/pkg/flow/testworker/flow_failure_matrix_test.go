package testworker

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateworkflowsteps"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (e *FlowTestSuite) setupLifecycleTest(ctx context.Context, ownerID, ownerType string, steps []app.WorkflowStep) (*app.Workflow, string) {
	stepQueue := e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-steps")
	e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-step-groups")
	e.createTestQueue(ctx, ownerID, ownerType, "install-signals")
	e.createTestQueue(ctx, ownerID, ownerType, "install-generate-steps")

	flw := app.Workflow{
		OwnerID:   ownerID,
		OwnerType: ownerType,
		Type:      "test_flow",
		Status:    app.NewCompositeStatus(ctx, app.StatusPending),
	}
	res := e.service.DB.WithContext(ctx).Create(&flw)
	require.Nil(e.T(), res.Error)

	e.createTestSteps(ctx, &flw, steps)
	return &flw, stepQueue.ID
}

// Validate only auto-resolves queue names for real owner types, so set all of them explicitly.
func (e *FlowTestSuite) enqueueLifecycleFlow(ctx context.Context, queueID string, flw *app.Workflow, ownerID, ownerType string) {
	resp, err := e.service.QueueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queueID,
		Signal: &executeflow.Signal{
			WorkflowID:             flw.ID,
			StepGroupQueueName:     "install-workflow-step-groups",
			StepQueueName:          "install-workflow-steps",
			StepTargetQueueName:    "install-signals",
			GenerateStepsQueueName: "install-generate-steps",
			OwnerID:                ownerID,
			OwnerType:              ownerType,
		},
		OwnerID:   flw.ID,
		OwnerType: "install_workflows",
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), resp)
}

func (e *FlowTestSuite) seedDeployTarget(ctx context.Context, status app.InstallDeployStatus) *app.InstallDeploy {
	orgID, err := cctx.OrgIDFromContext(ctx)
	require.NoError(e.T(), err)

	deploy := app.InstallDeploy{
		OrgID:              orgID,
		CreatedByID:        generics.GetFakeObj[string](),
		ComponentBuildID:   generics.GetFakeObj[string](),
		InstallComponentID: generics.GetFakeObj[string](),
		Status:             status,
		StatusDescription:  string(status),
	}
	// The fake FK targets don't exist; replica role skips constraint triggers.
	tx := e.service.DB.WithContext(ctx).Begin()
	require.NoError(e.T(), tx.Error)
	require.NoError(e.T(), tx.Exec("SET LOCAL session_replication_role = replica").Error)
	require.NoError(e.T(), tx.Omit(clause.Associations).Create(&deploy).Error)
	require.NoError(e.T(), tx.Commit().Error)
	return &deploy
}

func (e *FlowTestSuite) getDeploy(ctx context.Context, id string) *app.InstallDeploy {
	var deploy app.InstallDeploy
	res := e.service.DB.WithContext(ctx).First(&deploy, "id = ?", id)
	require.NoError(e.T(), res.Error)
	return &deploy
}

// statusMatrix asserts a step's combined lifecycle state; zero-valued fields are skipped.
type statusMatrix struct {
	Step     app.Status
	Workflow app.Status
	Run      app.Status
	Target   app.Status
}

func (e *FlowTestSuite) assertStatusMatrix(ctx context.Context, stepID string, want statusMatrix) {
	if want.Step != "" {
		e.waitForStepStatus(ctx, stepID, want.Step)
	}
	if want.Workflow != "" {
		step := e.getStep(ctx, stepID)
		e.waitForWorkflowStatus(ctx, step.InstallWorkflowID, want.Workflow)
	}
	if want.Run != "" {
		step := e.getStep(ctx, stepID)
		require.Eventually(e.T(), func() bool {
			runs := e.getWorkflowRuns(ctx, step.InstallWorkflowID)
			return len(runs) > 0 && runs[len(runs)-1].Status.Status == want.Run
		}, pollTimeout, pollInterval, "latest run did not reach status %s", want.Run)
	}
	if want.Target != "" {
		step := e.getStep(ctx, stepID)
		require.Equal(e.T(), string(app.WorkflowStepTargetTypeInstallDeploys), step.StepTargetType,
			"target assertions only support install_deploys")
		require.Eventually(e.T(), func() bool {
			return e.getDeploy(ctx, step.StepTargetID).Status == app.InstallDeployStatus(want.Target)
		}, pollTimeout, pollInterval, "target did not reach status %s", want.Target)
	}
}

func failingTargetedStep(name string, sig signaldb.SignalData, deploy *app.InstallDeploy) app.WorkflowStep {
	step := app.WorkflowStep{
		Name:          name,
		Idx:           100,
		GroupIdx:      1,
		ExecutionType: app.WorkflowStepExecutionTypeSystem,
		Retryable:     true,
		Skippable:     true,
		QueueSignal:   &sig,
	}
	if deploy != nil {
		step.StepTargetType = string(app.WorkflowStepTargetTypeInstallDeploys)
		step.StepTargetID = deploy.ID
	}
	return step
}

// Generated workflows materialize all steps up front; later steps stay pending while the first runs.
func (e *FlowTestSuite) TestGeneratedStepsStartPending() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	workflowType := app.WorkflowType("test_generated_steps_pending")

	generateworkflowsteps.RegisterGenerators(ownerType, func() map[app.WorkflowType]flow.WorkflowStepGenerator {
		return map[app.WorkflowType]flow.WorkflowStepGenerator{
			workflowType: func(workflow.Context, *app.Workflow) (*app.GenerateStepsResult, error) {
				return &app.GenerateStepsResult{
					Groups: []*app.WorkflowStepGroup{
						{GroupIdx: 1, Status: app.CompositeStatus{Status: app.StatusPending}},
						{GroupIdx: 2, Status: app.CompositeStatus{Status: app.StatusPending}},
						{GroupIdx: 3, Status: app.CompositeStatus{Status: app.StatusPending}},
					},
					Steps: []*app.WorkflowStep{
						{
							Name:          "blocking-first-step",
							Idx:           100,
							GroupIdx:      1,
							ExecutionType: app.WorkflowStepExecutionTypeSystem,
							Status:        app.CompositeStatus{Status: app.StatusPending},
							QueueSignal:   &signaldb.SignalData{Signal: &CancellableTestSignal{}},
						},
						{
							Name:          "second-step",
							Idx:           200,
							GroupIdx:      2,
							ExecutionType: app.WorkflowStepExecutionTypeSystem,
							Status:        app.CompositeStatus{Status: app.StatusPending},
							QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
						},
						{
							Name:          "third-step",
							Idx:           300,
							GroupIdx:      3,
							ExecutionType: app.WorkflowStepExecutionTypeSystem,
							Status:        app.CompositeStatus{Status: app.StatusPending},
							QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
						},
					},
				}, nil
			},
		}
	})

	workflowQueue := e.createTestQueue(ctx, ownerID, ownerType, "install-workflows")
	e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-steps")
	e.createTestQueue(ctx, ownerID, ownerType, "install-workflow-step-groups")
	e.createTestQueue(ctx, ownerID, ownerType, "install-signals")
	e.createTestQueue(ctx, ownerID, ownerType, "install-generate-steps")

	flw := e.createTestWorkflow(ctx, ownerID, ownerType, workflowType, &generateworkflowsteps.Signal{})
	e.enqueueLifecycleFlow(ctx, workflowQueue.ID, flw, ownerID, ownerType)

	e.waitForStepInProgress(ctx, flw.ID, "blocking-first-step")
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	require.Len(e.T(), steps, 3, "all steps must be generated up front")
	for _, s := range steps {
		switch s.Name {
		case "blocking-first-step":
			require.Equal(e.T(), app.StatusInProgress, s.Status.Status)
		default:
			require.Equal(e.T(), app.StatusPending, s.Status.Status,
				"%s must stay pending while the first step runs", s.Name)
		}
	}

	_, err := e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{InstallWorkflowID: flw.ID})
	require.NoError(e.T(), err)
	e.waitForWorkflowTerminal(ctx, flw.ID)
	e.assertTemporalDrained(ctx, flw.ID)
}

// Exhausted failure stops the workflow; only the runner ever writes success/failure target statuses.
func (e *FlowTestSuite) TestFailStopTargetUntouched() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	deploy := e.seedDeployTarget(ctx, app.InstallDeployStatusActive)

	steps := []app.WorkflowStep{
		failingTargetedStep("fail-stop", signaldb.SignalData{Signal: &SkippableFailSignal{}}, deploy),
		{
			Name:          "never-runs",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)
	e.assertStatusMatrix(ctx, steps[0].ID, statusMatrix{
		Step:     app.StatusError,
		Workflow: app.StatusError,
		Run:      app.StatusError,
		Target:   app.Status(app.InstallDeployStatusActive),
	})
	for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
		if s.Name == "never-runs" {
			require.Equal(e.T(), app.StatusNotAttempted, s.Status.Status)
		}
	}
	e.assertTemporalDrained(ctx, flw.ID)
}

// Auto-budget exhaustion parks the step; workflow failed-pending-retry, target untouched.
func (e *FlowTestSuite) TestFailParkTargetUntouched() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	deploy := e.seedDeployTarget(ctx, app.InstallDeployStatusActive)

	steps := []app.WorkflowStep{
		failingTargetedStep("fail-park", signaldb.SignalData{Signal: &ManualRetrySignal{}}, deploy),
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)
	step := e.getStep(ctx, steps[0].ID)
	require.Equal(e.T(), app.StatusError, step.Status.Status)
	require.Equal(e.T(), directive.StepAwaitRetry, directive.Step(step.ResultDirective))
	e.assertStatusMatrix(ctx, steps[0].ID, statusMatrix{
		Workflow: app.StatusFailedPendingRetry,
		Target:   app.Status(app.InstallDeployStatusActive),
	})
}

// One parallel failure keeps the sibling's success and still fails the workflow.
func (e *FlowTestSuite) TestFailParallelPartial() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		{
			Name:          "parallel-fail",
			Idx:           100,
			GroupIdx:      1,
			GroupParallel: true,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SkippableFailSignal{}},
		},
		{
			Name:          "parallel-success",
			Idx:           150,
			GroupIdx:      1,
			GroupParallel: true,
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
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)
	for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
		switch s.Name {
		case "parallel-fail":
			require.Equal(e.T(), app.StatusError, s.Status.Status)
		case "parallel-success":
			require.Equal(e.T(), app.StatusSuccess, s.Status.Status)
		case "later-group":
			require.Equal(e.T(), app.StatusNotAttempted, s.Status.Status)
		}
	}
	e.assertTemporalDrained(ctx, flw.ID)
}

// Manual retry creates exactly one clone; original keeps discarded+retried; workflow completes.
func (e *FlowTestSuite) TestFailManualRetryOnce() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		failingTargetedStep("retry-me-once", signaldb.SignalData{Signal: &ManualRetrySignal{}}, nil),
		{
			Name:          "after-retry",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	_, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            steps[0].ID,
	})
	require.NoError(e.T(), err)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	original := e.getStep(ctx, steps[0].ID)
	require.True(e.T(), original.Retried, "original step must be marked retried")
	require.Equal(e.T(), app.StatusDiscarded, original.Status.Status)

	cloneCount := 0
	for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
		if s.Name == "retry-me-once" && s.ID != steps[0].ID {
			cloneCount++
			require.Equal(e.T(), 1, s.RetryIndex)
			require.Equal(e.T(), app.StatusSuccess, s.Status.Status)
		}
	}
	require.Equal(e.T(), 1, cloneCount, "manual retry must create exactly one clone")
	e.assertTemporalDrained(ctx, flw.ID)
}

// While a retry clone executes, the workflow reports in-progress, not failed-pending-retry.
func (e *FlowTestSuite) TestFailManualRetryRunsAsInProgress() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		failingTargetedStep("retry-then-block", signaldb.SignalData{Signal: &ManualRetryThenBlockSignal{}}, nil),
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	_, err := e.service.FlowClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            steps[0].ID,
	})
	require.NoError(e.T(), err)

	e.waitForStepInProgress(ctx, flw.ID, "retry-then-block")
	require.Equal(e.T(), app.StatusInProgress, e.getWorkflow(ctx, flw.ID).Status.Status,
		"workflow must report in-progress while the retry clone executes")

	_, err = e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{InstallWorkflowID: flw.ID})
	require.NoError(e.T(), err)
	e.waitForWorkflowTerminal(ctx, flw.ID)
	e.assertTemporalDrained(ctx, flw.ID)
}

// Skipping a parked step marks it user-skipped and the workflow resumes running.
func (e *FlowTestSuite) TestFailSkipThenContinue() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		failingTargetedStep("skip-me-after-failure", signaldb.SignalData{Signal: &ManualRetrySignal{}}, nil),
		{
			Name:          "blocking-after-skip",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &CancellableTestSignal{}},
		},
		{
			Name:          "final-step",
			Idx:           300,
			GroupIdx:      3,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	resp, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            steps[0].ID,
	})
	require.NoError(e.T(), err)
	require.True(e.T(), resp.Skippable)

	e.waitForStepStatus(ctx, steps[0].ID, app.StatusUserSkipped)
	blockingID := e.waitForStepInProgress(ctx, flw.ID, "blocking-after-skip")
	require.Equal(e.T(), app.StatusInProgress, e.getWorkflow(ctx, flw.ID).Status.Status,
		"workflow must report in-progress while executing the group after the skip")

	_, err = e.service.FlowClient.CancelStep(ctx, &flowclient.CancelStepRequest{
		InstallWorkflowID: flw.ID,
		StepID:            blockingID,
	})
	require.NoError(e.T(), err)
	e.waitForStepStatus(ctx, blockingID, app.StatusCancelled)
}

// Cancelling a parked step cancels step+target; main writes no terminal workflow status (see TestPinCancelStepTerminatesWorkflow).
func (e *FlowTestSuite) TestFailCancelWhileParked() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	deploy := e.seedDeployTarget(ctx, app.InstallDeployStatusError)

	steps := []app.WorkflowStep{
		failingTargetedStep("cancel-me-while-parked", signaldb.SignalData{Signal: &ManualRetrySignal{}}, deploy),
		{
			Name:          "later-group",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
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

// SkipOnFailure exhaustion marks the step failed (skipped_on_failure) and the workflow continues.
func (e *FlowTestSuite) TestSkipOnFailureContinuesWorkflow() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		{
			Name:          "skippable-failure",
			Idx:           100,
			GroupIdx:      1,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			SkipOnFailure: true,
			QueueSignal:   &signaldb.SignalData{Signal: &SkippableFailSignal{}},
		},
		{
			Name:          "after-skippable-failure",
			Idx:           200,
			GroupIdx:      2,
			ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
		},
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	step := e.getStep(ctx, steps[0].ID)
	require.Equal(e.T(), app.StatusError, step.Status.Status)
	require.Equal(e.T(), true, step.Status.Metadata["skipped_on_failure"])
	for _, s := range e.getStepsByWorkflow(ctx, flw.ID) {
		if s.Name == "after-skippable-failure" {
			require.Equal(e.T(), app.StatusSuccess, s.Status.Status,
				"the group after a skip-on-failure step must still run")
		}
	}
	e.assertTemporalDrained(ctx, flw.ID)
}

// The approve path never writes the deploy target; success statuses are the runner's job.
func (e *FlowTestSuite) TestApproveTargetUntouchedOnSuccess() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	deploy := e.seedDeployTarget(ctx, app.InstallDeployStatusActive)

	steps := []app.WorkflowStep{
		approvalStep("approve-with-target", 1, signaldb.SignalData{Signal: &ApprovalInnerSignal{}}),
	}
	steps[0].StepTargetType = string(app.WorkflowStepTargetTypeInstallDeploys)
	steps[0].StepTargetID = deploy.ID
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	approval := e.seedApproval(ctx, &steps[0])

	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[0].ID)
	e.respondApproval(ctx, flw, &steps[0], approval.ID, app.WorkflowStepApprovalResponseTypeApprove)

	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	e.assertStatusMatrix(ctx, steps[0].ID, statusMatrix{
		Target: app.Status(app.InstallDeployStatusActive),
	})
	e.assertTemporalDrained(ctx, flw.ID)
}
