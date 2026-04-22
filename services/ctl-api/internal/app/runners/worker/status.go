package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/processjobsignals"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (w *Workflows) updateStatus(ctx workflow.Context, runnerID string, status app.RunnerStatus, statusDescription string) {
	err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          runnerID,
		Status:            status,
		StatusDescription: statusDescription,
	})

	l := workflow.GetLogger(ctx)
	if err != nil {
		l.Error("unable to update runner status",
			zap.String("runner-id", runnerID),
			zap.Error(err))
		return
	}

	w.signalActiveJobsForRunner(ctx, runnerID, processjobsignals.ReasonRunnerStatusChanged)
}

// signalActiveJobsForRunner fetches the IDs of all queued/in-progress jobs for
// this runner and sends the ProcessJob wake-up signal to each. Errors are
// logged but never propagated — this is best-effort.
func (w *Workflows) signalActiveJobsForRunner(ctx workflow.Context, runnerID, reason string) {
	l := workflow.GetLogger(ctx)

	jobIDs, err := activities.AwaitGetActiveJobIDsForRunner(ctx, runnerID)
	if err != nil {
		l.Error("unable to fetch active job IDs for runner status signal",
			zap.String("runner-id", runnerID),
			zap.Error(err))
		return
	}

	for _, jobID := range jobIDs {
		wakeup := processjobsignals.WakeUp{Reason: reason}
		if err := workflow.SignalExternalWorkflow(ctx,
			processjobsignals.WorkflowID(jobID),
			"",
			processjobsignals.SignalName,
			wakeup,
		).Get(ctx, nil); err != nil {
			l.Debug("processjob signal skipped",
				zap.String("job-id", jobID),
				zap.String("runner-id", runnerID),
				zap.Error(err))
		}
	}
}

func (w *Workflows) updateJobStatus(ctx workflow.Context, jobID string, status app.RunnerJobStatus, statusDescription string) {
	err := activities.AwaitUpdateJobStatus(ctx, activities.UpdateJobStatusRequest{
		JobID:             jobID,
		Status:            status,
		StatusDescription: statusDescription,
	})

	statusactivities.AwaitUpdateRunnerJobStatusV2(ctx, statusactivities.UpdateRunnerJobStatusV2Request{
		RunnerJobID:       jobID,
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

	statusactivities.AwaitUpdateRunnerJobExecutionStatusV2(ctx, statusactivities.UpdateRunnerJobExecutionStatusV2Request{
		RunnerJobExecutionID: jobExecutionID,
		Status:               status,
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
