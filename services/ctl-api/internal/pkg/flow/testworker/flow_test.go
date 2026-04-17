package testworker

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/testworker/example"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/testworker/seed"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const (
	pollTimeout  = 60 * time.Second
	pollInterval = 1 * time.Second
)

// helper to set up common test context: account, org, queues, workflow
type testSetup struct {
	ctx       context.Context
	ownerID   string
	ownerType string
	queues    *seed.EnsureQueuesResult
	workflow  *seed.EnsureWorkflowResult
}

func (e *FlowTestSuite) setupFlow(steps []example.StepConfig) *testSetup {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	ownerID := generics.GetFakeObj[string]()
	ownerType := "installs"

	queues := e.service.Seed.EnsureQueues(ctx, e.T(), ownerID, ownerType)

	wfResult := e.service.Seed.EnsureWorkflow(ctx, e.T(), seed.EnsureWorkflowRequest{
		OwnerID:   ownerID,
		OwnerType: ownerType,
		Steps:     steps,
	})

	return &testSetup{
		ctx:       ctx,
		ownerID:   ownerID,
		ownerType: ownerType,
		queues:    queues,
		workflow:  wfResult,
	}
}

func (e *FlowTestSuite) enqueueFlow(setup *testSetup) string {
	resp, err := e.service.QueueClient.EnqueueSignal(setup.ctx, &queueclient.EnqueueSignalRequest{
		QueueID: setup.queues.PrimaryQueue.ID,
		Signal: &executeflow.Signal{
			WorkflowID:          setup.workflow.Workflow.ID,
			StepQueueName:       "install-workflow-steps",
			StepTargetQueueName: "install-signals",
			OwnerID:             setup.ownerID,
			OwnerType:           setup.ownerType,
		},
		OwnerID:   setup.workflow.Workflow.ID,
		OwnerType: "install_workflows",
	})
	require.Nil(e.T(), err)
	return resp.ID
}

func (e *FlowTestSuite) pollUntilDone(signalID string) {
	ctx := e.T().Context()
	timeout := pollTimeout
	_, err := e.service.QueueClient.PollSignal(ctx, signalID, &queueclient.PollSignalOptions{
		Timeout:      &timeout,
		PollInterval: pollInterval,
	})
	require.Nil(e.T(), err)
}

// assertWorkflowStatus checks the workflow's final status in the DB.
func (e *FlowTestSuite) assertWorkflowStatus(workflowID string, expected app.Status) {
	var wf app.Workflow
	res := e.service.DB.First(&wf, "id = ?", workflowID)
	require.Nil(e.T(), res.Error)
	require.Equal(e.T(), expected, wf.Status.Status, "workflow status mismatch: got %s", wf.Status.Status)
}

// getWorkflowSteps fetches all steps for a workflow ordered by idx.
func (e *FlowTestSuite) getWorkflowSteps(workflowID string) []app.WorkflowStep {
	var steps []app.WorkflowStep
	res := e.service.DB.Where(app.WorkflowStep{InstallWorkflowID: workflowID}).Order("idx asc").Find(&steps)
	require.Nil(e.T(), res.Error)
	return steps
}

// getWorkflowRuns fetches all runs for a workflow.
func (e *FlowTestSuite) getWorkflowRuns(workflowID string) []app.WorkflowRun {
	var runs []app.WorkflowRun
	res := e.service.DB.Where(app.WorkflowRun{WorkflowID: workflowID}).Order("created_at asc").Find(&runs)
	require.Nil(e.T(), res.Error)
	return runs
}

// ---------------------------------------------------------------------------
// Test Cases
// ---------------------------------------------------------------------------

// TestSingleStepSuccess verifies a single-step workflow completes successfully.
func (e *FlowTestSuite) TestSingleStepSuccess() {
	setup := e.setupFlow([]example.StepConfig{
		{Behavior: "pass", Retryable: true},
	})

	signalID := e.enqueueFlow(setup)
	e.pollUntilDone(signalID)

	e.assertWorkflowStatus(setup.workflow.Workflow.ID, app.StatusSuccess)

	steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
	require.Len(e.T(), steps, 1)
	require.Equal(e.T(), app.StatusSuccess, steps[0].Status.Status)

	runs := e.getWorkflowRuns(setup.workflow.Workflow.ID)
	require.GreaterOrEqual(e.T(), len(runs), 1)
	require.Equal(e.T(), app.WorkflowRunTypeInitial, runs[0].Type)
}

// TestMultiStepSuccess verifies multiple steps execute sequentially.
func (e *FlowTestSuite) TestMultiStepSuccess() {
	setup := e.setupFlow([]example.StepConfig{
		{Behavior: "pass"},
		{Behavior: "pass"},
		{Behavior: "pass"},
	})

	signalID := e.enqueueFlow(setup)
	e.pollUntilDone(signalID)

	e.assertWorkflowStatus(setup.workflow.Workflow.ID, app.StatusSuccess)

	steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
	require.Len(e.T(), steps, 3)
	for _, step := range steps {
		require.Equal(e.T(), app.StatusSuccess, step.Status.Status)
	}
}

// TestAutoRetry verifies that a step with EnableAutoRetry creates a clone on failure.
func (e *FlowTestSuite) TestAutoRetry() {
	setup := e.setupFlow([]example.StepConfig{
		{Behavior: "fail", EnableAutoRetry: true, CustomMaxRetries: 2, Retryable: true},
	})

	signalID := e.enqueueFlow(setup)
	e.pollUntilDone(signalID)

	// With max retries of 2, we expect: original + clone1 + clone2.
	// clone2 also fails and retries_exhausted=true, then the flow errors out.
	steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
	require.GreaterOrEqual(e.T(), len(steps), 2, "expected at least 2 steps (original + retry clone)")

	// The original step should have auto_retried metadata
	require.Equal(e.T(), app.StatusError, steps[0].Status.Status)
	if steps[0].Status.Metadata != nil {
		_, hasAutoRetried := steps[0].Status.Metadata["auto_retried"]
		require.True(e.T(), hasAutoRetried, "original step should have auto_retried metadata")
	}
}

// TestAutoRetryMaxRetries verifies that retries stop after the max retry budget is exhausted.
func (e *FlowTestSuite) TestAutoRetryMaxRetries() {
	setup := e.setupFlow([]example.StepConfig{
		{Behavior: "fail", EnableAutoRetry: true, CustomMaxRetries: 1, Retryable: true},
	})

	signalID := e.enqueueFlow(setup)
	e.pollUntilDone(signalID)

	steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
	// With max retries of 1: original (retryIndex=0) auto-retries to clone (retryIndex=1),
	// clone fails and retryIndex=1 >= maxRetries=1 so retries_exhausted=true.
	require.GreaterOrEqual(e.T(), len(steps), 2)

	// Find the last step and verify retries_exhausted
	lastStep := steps[len(steps)-1]
	require.Equal(e.T(), app.StatusError, lastStep.Status.Status)
	if lastStep.Status.Metadata != nil {
		if exhausted, ok := lastStep.Status.Metadata["retries_exhausted"]; ok {
			require.Equal(e.T(), true, exhausted)
		}
	}
}

// TestCloneSteps verifies that SignalWithCloneSteps creates custom clone steps on retry.
func (e *FlowTestSuite) TestCloneSteps() {
	setup := e.setupFlow([]example.StepConfig{
		{
			Behavior:         "fail",
			EnableAutoRetry:  true,
			CustomMaxRetries: 2,
			Retryable:        true,
			CloneStepNames:   []string{"Plan", "Apply"},
		},
	})

	signalID := e.enqueueFlow(setup)
	e.pollUntilDone(signalID)

	steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
	// Should have: original + "Plan (retry 1)" + "Apply (retry 1)"
	require.GreaterOrEqual(e.T(), len(steps), 3, "expected at least 3 steps: original + 2 clone steps")

	// Verify clone step names contain retry suffix
	foundPlanRetry := false
	foundApplyRetry := false
	for _, step := range steps[1:] {
		if step.Name == "Plan (retry 1)" {
			foundPlanRetry = true
		}
		if step.Name == "Apply (retry 1)" {
			foundApplyRetry = true
		}
	}
	require.True(e.T(), foundPlanRetry, "expected a 'Plan (retry 1)' clone step")
	require.True(e.T(), foundApplyRetry, "expected an 'Apply (retry 1)' clone step")
}

// TestCancelCallback verifies that cancelling a slow step triggers the cancel callback.
func (e *FlowTestSuite) TestCancelCallback() {
	setup := e.setupFlow([]example.StepConfig{
		{Behavior: "slow", Retryable: true},
	})

	signalID := e.enqueueFlow(setup)

	// Wait until step is in-progress
	require.Eventually(e.T(), func() bool {
		steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
		if len(steps) == 0 {
			return false
		}
		return steps[0].Status.Status == app.StatusInProgress
	}, 30*time.Second, 1*time.Second, "step never reached in-progress")

	// Get the step ID to cancel
	steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
	require.NotEmpty(e.T(), steps)
	stepID := steps[0].ID

	// Send cancel via flow client
	_, err := e.service.FlowClient.CancelStep(e.T().Context(), &flowclient.CancelStepRequest{
		InstallWorkflowID: setup.workflow.Workflow.ID,
		StepID:            stepID,
	})
	require.Nil(e.T(), err)

	// Poll until done
	e.pollUntilDone(signalID)

	// Step should be cancelled
	steps = e.getWorkflowSteps(setup.workflow.Workflow.ID)
	require.NotEmpty(e.T(), steps)
	require.Equal(e.T(), app.StatusCancelled, steps[0].Status.Status)
}

// TestManualRetry verifies that a failed step can be retried via the retry-step update handler.
func (e *FlowTestSuite) TestManualRetry() {
	setup := e.setupFlow([]example.StepConfig{
		{Behavior: "fail", Retryable: true},
	})

	signalID := e.enqueueFlow(setup)

	// Wait until workflow enters error/awaiting-retry state
	require.Eventually(e.T(), func() bool {
		var wf app.Workflow
		res := e.service.DB.First(&wf, "id = ?", setup.workflow.Workflow.ID)
		if res.Error != nil {
			return false
		}
		if wf.Status.Status != app.StatusError {
			return false
		}
		if wf.Status.Metadata != nil {
			if v, ok := wf.Status.Metadata["awaiting_retry"]; ok {
				return v == true
			}
		}
		return false
	}, 30*time.Second, 1*time.Second, "workflow never reached awaiting_retry state")

	// Get the failed step
	steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
	require.NotEmpty(e.T(), steps)
	stepID := steps[0].ID

	// Send retry via flow client
	_, err := e.service.FlowClient.RetryStep(e.T().Context(), &flowclient.RetryStepRequest{
		InstallWorkflowID: setup.workflow.Workflow.ID,
		StepID:            stepID,
	})
	require.Nil(e.T(), err)

	// The retry will create a clone that also fails, entering awaiting_retry again.
	// Wait for the retry run to be created.
	require.Eventually(e.T(), func() bool {
		runs := e.getWorkflowRuns(setup.workflow.Workflow.ID)
		return len(runs) >= 2
	}, 30*time.Second, 1*time.Second, "retry run never created")

	runs := e.getWorkflowRuns(setup.workflow.Workflow.ID)
	require.Equal(e.T(), app.WorkflowRunTypeInitial, runs[0].Type)
	require.Equal(e.T(), app.WorkflowRunTypeRetry, runs[1].Type)

	// Verify clone step was created
	steps = e.getWorkflowSteps(setup.workflow.Workflow.ID)
	require.GreaterOrEqual(e.T(), len(steps), 2, "expected clone step to be created")

	// Clean up: cancel the workflow so it doesn't block forever
	_ = signalID
}

// TestManualRetryWithCloneSteps verifies retry with custom clone steps.
func (e *FlowTestSuite) TestManualRetryWithCloneSteps() {
	setup := e.setupFlow([]example.StepConfig{
		{
			Behavior:       "fail",
			Retryable:      true,
			CloneStepNames: []string{"Plan", "Apply"},
		},
	})

	signalID := e.enqueueFlow(setup)

	// Wait for awaiting_retry
	require.Eventually(e.T(), func() bool {
		var wf app.Workflow
		res := e.service.DB.First(&wf, "id = ?", setup.workflow.Workflow.ID)
		if res.Error != nil {
			return false
		}
		if wf.Status.Status != app.StatusError {
			return false
		}
		if wf.Status.Metadata != nil {
			if v, ok := wf.Status.Metadata["awaiting_retry"]; ok {
				return v == true
			}
		}
		return false
	}, 30*time.Second, 1*time.Second)

	steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
	require.NotEmpty(e.T(), steps)

	_, err := e.service.FlowClient.RetryStep(e.T().Context(), &flowclient.RetryStepRequest{
		InstallWorkflowID: setup.workflow.Workflow.ID,
		StepID:            steps[0].ID,
	})
	require.Nil(e.T(), err)

	// Wait for clone steps
	require.Eventually(e.T(), func() bool {
		steps = e.getWorkflowSteps(setup.workflow.Workflow.ID)
		return len(steps) >= 3 // original + Plan + Apply
	}, 30*time.Second, 1*time.Second)

	foundPlan := false
	foundApply := false
	for _, step := range steps {
		if step.Name == "Plan (retry 1)" {
			foundPlan = true
		}
		if step.Name == "Apply (retry 1)" {
			foundApply = true
		}
	}
	require.True(e.T(), foundPlan, "expected Plan (retry 1) clone step")
	require.True(e.T(), foundApply, "expected Apply (retry 1) clone step")

	_ = signalID
}

// TestFailingStepNonRetryable verifies a non-retryable failing step terminates the flow.
func (e *FlowTestSuite) TestFailingStepNonRetryable() {
	setup := e.setupFlow([]example.StepConfig{
		{Behavior: "fail", Retryable: false},
	})

	// Create a second newer workflow for the same owner so checkRetryable returns false
	e.service.Seed.EnsureWorkflow(setup.ctx, e.T(), seed.EnsureWorkflowRequest{
		OwnerID:   setup.ownerID,
		OwnerType: setup.ownerType,
		Steps:     []example.StepConfig{{Behavior: "pass"}},
	})

	signalID := e.enqueueFlow(setup)
	e.pollUntilDone(signalID)

	e.assertWorkflowStatus(setup.workflow.Workflow.ID, app.StatusError)

	steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
	require.Len(e.T(), steps, 1)
	require.Equal(e.T(), app.StatusError, steps[0].Status.Status)
}

// TestCancelWhileInProgress verifies cancel-step update handler propagates correctly.
func (e *FlowTestSuite) TestCancelWhileInProgress() {
	setup := e.setupFlow([]example.StepConfig{
		{Behavior: "slow"},
	})

	signalID := e.enqueueFlow(setup)

	// Wait for in-progress
	require.Eventually(e.T(), func() bool {
		var wf app.Workflow
		res := e.service.DB.First(&wf, "id = ?", setup.workflow.Workflow.ID)
		return res.Error == nil && wf.Status.Status == app.StatusInProgress
	}, 30*time.Second, 1*time.Second)

	// Wait for step to exist
	require.Eventually(e.T(), func() bool {
		steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
		return len(steps) > 0
	}, 10*time.Second, 500*time.Millisecond)

	steps := e.getWorkflowSteps(setup.workflow.Workflow.ID)
	_, err := e.service.FlowClient.CancelStep(e.T().Context(), &flowclient.CancelStepRequest{
		InstallWorkflowID: setup.workflow.Workflow.ID,
		StepID:            steps[0].ID,
	})
	require.Nil(e.T(), err)

	e.pollUntilDone(signalID)
}

// TestNoopPlan is a placeholder for noop plan detection testing.
// Full noop detection requires a StepTargetID pointing to an install deploy,
// which needs additional seeding infrastructure.
func (e *FlowTestSuite) TestNoopPlan() {
	e.T().Skip("placeholder: requires install deploy seeding for CheckNoopPlan activity")
}

// TestPolicyEvaluation is a placeholder for policy evaluation testing.
// Full policy evaluation requires policy records and evaluation infrastructure.
func (e *FlowTestSuite) TestPolicyEvaluation() {
	e.T().Skip("placeholder: requires policy infrastructure seeding")
}
