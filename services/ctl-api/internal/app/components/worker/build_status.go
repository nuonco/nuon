package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (w *Workflows) updateBuildStatus(ctx workflow.Context, bldID string, status app.ComponentBuildStatus, statusDescription string) error {
	l := workflow.GetLogger(ctx)
	err := activities.AwaitUpdateBuildStatus(ctx, activities.UpdateBuildStatus{
		BuildID:           bldID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err != nil {
		l.Error("unable to update build status",
			zap.String("build-id", bldID),
			zap.Error(err))
		return fmt.Errorf("unable to update build status: %w", err)
	}

	err = statusactivities.AwaitUpdateBuildStatusV2(ctx, statusactivities.UpdateBuildStatusV2{
		BuildID:           bldID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err != nil {
		l.Error("unable to update build status v2",
			zap.String("build-id", bldID),
			zap.Error(err))
		return fmt.Errorf("unable to update build status history: %w", err)
	}
	return nil
}

func (w *Workflows) updateBuildSignatureVerification(ctx workflow.Context, bldID, jobID string, pullPlan *plantypes.ContainerImagePullPlan, jobSucceeded bool) {
	err := activities.AwaitUpdateBuildSignatureVerification(ctx, &activities.UpdateBuildSignatureVerificationRequest{
		BuildID:      bldID,
		JobID:        jobID,
		Required:     pullPlan.Verification != nil && pullPlan.Verification.RequireSignature,
		JobSucceeded: jobSucceeded,
	})
	if err != nil {
		workflow.GetLogger(ctx).Error("unable to update build signature verification",
			zap.String("build-id", bldID),
			zap.Error(err))
	}
}

func (w *Workflows) failRunnerJob(ctx workflow.Context, jobID, description string) {
	_ = activities.AwaitUpdateJobStatus(ctx, &activities.UpdateJobStatusRequest{
		JobID:             jobID,
		Status:            app.RunnerJobStatusFailed,
		StatusDescription: description,
	})
}

func (w *Workflows) failQueuedRunnerJob(ctx workflow.Context, jobID, description string) error {
	return activities.AwaitUpdateJobStatus(ctx, &activities.UpdateJobStatusRequest{
		JobID:             jobID,
		Status:            app.RunnerJobStatusFailed,
		StatusDescription: description,
		ExpectedStatus:    app.RunnerJobStatusQueued,
	})
}
