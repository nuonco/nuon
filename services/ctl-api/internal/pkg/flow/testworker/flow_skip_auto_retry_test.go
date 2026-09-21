package testworker

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (e *FlowTestSuite) seedDeployJobWithSkipAutoRetry(ctx context.Context) string {
	orgID, err := cctx.OrgIDFromContext(ctx)
	require.Nil(e.T(), err)
	accountID, err := cctx.AccountIDFromContext(ctx)
	require.Nil(e.T(), err)

	targetID := domains.NewDeployID()
	now := time.Now()
	jobID := domains.NewRunnerJobID()
	compositeError := &compositeerrors.CompositeErrorData{
		Version:  compositeerrors.SchemaVersion,
		Type:     "terraform.aws_permission",
		Severity: compositeerrors.SeverityError,
		Message:  "AccessDenied: not authorized to perform s3:CreateBucket",
		Hints: compositeerrors.Hints{
			compositeerrors.HintSkipAutoRetry: "true",
		},
	}
	job := map[string]any{
		"id":                 jobID,
		"created_by_id":      accountID,
		"created_at":         now,
		"updated_at":         now,
		"deleted_at":         0,
		"org_id":             orgID,
		"owner_id":           targetID,
		"owner_type":         "install_deploys",
		"queue_timeout":      0,
		"available_timeout":  0,
		"execution_timeout":  0,
		"overall_timeout":    0,
		"max_executions":     1,
		"status":             app.RunnerJobStatusFailed,
		"status_description": "permission denied",
		"status_v2":          app.NewCompositeStatus(ctx, app.Status(app.RunnerJobStatusFailed)),
		"type":               app.RunnerJobTypeTerraformDeploy,
		"group":              app.RunnerJobGroupDeploy,
		"operation":          app.RunnerJobOperationTypeApplyPlan,
		"composite_error":    compositeError,
	}
	require.NoError(e.T(), e.service.DB.WithContext(ctx).Table("runner_jobs").Create(job).Error)
	execution := app.RunnerJobExecution{RunnerJobID: jobID, Status: app.RunnerJobExecutionStatusFailed}
	require.NoError(e.T(), e.service.DB.WithContext(ctx).Create(&execution).Error)
	result := app.RunnerJobExecutionResult{RunnerJobExecutionID: execution.ID, CompositeError: compositeError}
	require.NoError(e.T(), e.service.DB.WithContext(ctx).Create(&result).Error)
	return targetID
}

// TestSkipAutoRetryParksStepForManualRetry verifies that when a failed step's
// target carries a composite error with the skip_auto_retry hint, the step is
// parked for manual retry on the FIRST failure, the workflow reaches
// StatusFailedPendingRetry and no auto-retry clones are created, instead of
// burning through the signal's auto-retry budget.
func (e *FlowTestSuite) TestSkipAutoRetryParksStepForManualRetry() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()

	deployID := e.seedDeployJobWithSkipAutoRetry(ctx)

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "skip-auto-retry-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:      true,
			StepTargetType: string(app.WorkflowStepTargetTypeInstallDeploys),
			StepTargetID:   deployID,
			QueueSignal:    &signaldb.SignalData{Signal: &AutoRetrySignal{}}},
	})

	e.enqueueFlow(ctx, queueID, flw, ownerID, ownerType)

	// The signal always fails and normally auto-retries 3 times, but the
	// skip_auto_retry hint forces an immediate park for manual retry.
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusFailedPendingRetry)

	// No auto-retry clones should have been created.
	steps := e.getStepsByWorkflow(ctx, flw.ID)
	require.Len(e.T(), steps, 1,
		"expected exactly 1 step (no auto-retry clones), got %d", len(steps))

	step := steps[0]
	require.Equal(e.T(), app.StatusError, step.Status.Status)
	require.Equal(e.T(), string(directive.StepAwaitRetry), step.ResultDirective,
		"step should be parked with the await-retry directive")
	require.Equal(e.T(), true, step.Status.Metadata["skip_auto_retry"],
		"step metadata should record skip_auto_retry")
}
