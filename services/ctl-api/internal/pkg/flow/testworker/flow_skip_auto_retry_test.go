package testworker

import (
	"context"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// seedDeployWithSkipAutoRetry creates a minimal InstallDeploy row carrying a
// composite error whose hints request skip_auto_retry, mirroring what the
// runner-result chokepoint writes for e.g. a missing IAM permission.
func (e *FlowTestSuite) seedDeployWithSkipAutoRetry(ctx context.Context) *app.InstallDeploy {
	orgID, err := cctx.OrgIDFromContext(ctx)
	require.Nil(e.T(), err)

	deploy := app.InstallDeploy{
		OrgID:              orgID,
		CreatedByID:        generics.GetFakeObj[string](),
		ComponentBuildID:   generics.GetFakeObj[string](),
		InstallComponentID: generics.GetFakeObj[string](),
		Status:             app.InstallDeployStatus(app.StatusError),
		StatusDescription:  "permission denied",
		CompositeError: &compositeerrors.CompositeErrorData{
			Version:  compositeerrors.SchemaVersion,
			Type:     "terraform.aws_permission",
			Severity: compositeerrors.SeverityError,
			Message:  "AccessDenied: not authorized to perform s3:CreateBucket",
			Hints: compositeerrors.Hints{
				compositeerrors.HintSkipAutoRetry: "true",
			},
		},
	}
	tx := e.service.DB.WithContext(ctx).Begin()
	require.NoError(e.T(), tx.Error)
	require.NoError(e.T(), tx.Exec("SET LOCAL session_replication_role = replica").Error)
	require.NoError(e.T(), tx.Omit(clause.Associations).Create(&deploy).Error)
	require.NoError(e.T(), tx.Commit().Error)
	return &deploy
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

	deploy := e.seedDeployWithSkipAutoRetry(ctx)

	flw, queueID := e.setupFlowTest(ctx, ownerID, ownerType, []app.WorkflowStep{
		{Name: "skip-auto-retry-step", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			Retryable:      true,
			StepTargetType: string(app.WorkflowStepTargetTypeInstallDeploys),
			StepTargetID:   deploy.ID,
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
