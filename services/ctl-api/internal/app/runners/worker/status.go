package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

func (w *Workflows) updateStatus(ctx workflow.Context, runnerID string, status app.RunnerStatus, statusDescription string) {
	err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          runnerID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update runner status",
		zap.String("runner-id", runnerID),
		zap.Error(err))
}

func (w *Workflows) updateJobStatus(ctx workflow.Context, jobID string, status app.RunnerJobStatus, statusDescription string) {
	err := activities.AwaitUpdateJobStatus(ctx, activities.UpdateJobStatusRequest{
		JobID:             jobID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update runner job status",
		zap.String("runner-job-id", jobID),
		zap.Error(err))
}

func (w *Workflows) updateJobExecutionStatus(ctx workflow.Context, jobExecutionID string, status app.RunnerJobExecutionStatus) {
	err := activities.AwaitUpdateJobExecutionStatus(ctx, activities.UpdateJobExecutionStatusRequest{
		JobExecutionID: jobExecutionID,
		Status:         status,
	})
	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update runner job execution status",
		zap.String("runner-job-execution id", jobExecutionID),
		zap.Error(err))
}

func (w *Workflows) updateOperationStatus(ctx workflow.Context, opID string, status app.RunnerOperationStatus) {
	err := activities.AwaitUpdateOperation(ctx, activities.UpdateOperationRequest{
		OperationID: opID,
		Status:      status,
	})
	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update runner operation execution status",
		zap.String("runner-operation id", opID),
		zap.Error(err))
}
