package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

const (
	defaultJobPollPeriod       time.Duration = time.Second * 1
	defaultAvailablePollPeriod time.Duration = time.Second * 1
)

// this function is the most core part of the runner job system, it's responsible for a.) marking a job as available and
// then b.) waiting until it is picked up by a runner (ie: an execution exists) and then c.) finished.
//
// it is responsible for updating the runner job with each state, and in some cases the runner job execution.
func (w *Workflows) processJobExecution(ctx workflow.Context, job *app.RunnerJob) (bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	availableStart := workflow.Now(ctx)
	l.Info("marking job as available for runner to pick up")
	w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusAvailable, "waiting for runner to reserve job")

	start := workflow.Now(ctx)
	overallTimeout := job.CreatedAt.Add(job.OverallTimeout)
	availableTimeout := start.Add(job.AvailableTimeout)

	var (
		jobExecution      *app.RunnerJobExecution
		jobExecutionFound bool
	)

	// poll until the job is picked up, and an execution exists
	for !jobExecutionFound {
		workflow.Sleep(ctx, defaultAvailablePollPeriod)

		now := workflow.Now(ctx)
		if now.After(overallTimeout) {
			l.Error("overall job timeout reached")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "overall timeout")
			return false, nil
		}

		if now.After(availableTimeout) {
			l.Error("timeout waiting for job to be picked up")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "timeout waiting for runner to pick up job")
			return true, nil
		}

		// if the runner is deemed unhealthy, the job execution is marked as unknown, and the job is marked as
		// not attempted with the correct status, this is retryable.
		runnerStatus, err := w.getRunnerStatus(ctx, job.RunnerID)
		if err != nil {
			return false, err
		}
		if runnerStatus != app.RunnerStatusActive {
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusUnknown, "runner is unhealthy")
			return true, nil
		}

		jobStatus, err := activities.AwaitGetJobStatusByID(ctx, job.ID)
		if err != nil {
			return false, err
		}
		if jobStatus == app.RunnerJobStatusCancelled {
			return true, nil
		}

		// fetch latest job execution
		jobExecutionResp, err := activities.AwaitGetLatestJobExecution(ctx, activities.GetLatestJobExecutionRequest{
			JobID:       job.ID,
			AvailableAt: availableStart,
		})
		if err != nil {
			return false, fmt.Errorf("error fetching latest job execution: %w", err)
		}

		jobExecutionFound = jobExecutionResp.Found
		jobExecution = jobExecutionResp.JobExecution
	}

	l.Info("job picked up by runner and is in progress")
	w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusInProgress, "job execution in progress")

	// poll the job execution, until it's completed
	executionTimeout := jobExecution.CreatedAt.Add(job.ExecutionTimeout)
	for {
		workflow.Sleep(ctx, defaultJobPollPeriod)

		now := workflow.Now(ctx)

		// when the overall timeout is hit, we mark both the runner job and execution as timed out
		// this is not retryable.
		if now.After(overallTimeout) {
			l.Error("overall timeout reached")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "overall timeout")
			w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusTimedOut)
			return false, nil
		}

		// when the execution timeout is hit, we mark both the runner job and execution as timed out
		// this is retryable
		if now.After(executionTimeout) {
			l.Error("execution timeout reached")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "execution timeout")
			w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusTimedOut)
			return true, nil
		}

		jobStatus, err := activities.AwaitGetJobStatusByID(ctx, job.ID)
		if err != nil {
			return false, err
		}
		if jobStatus == app.RunnerJobStatusCancelled {
			l.Error("job was cancelled")
			w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusCancelled)
			return true, nil
		}

		// if the runner is deemed unhealthy, the job execution is marked as unknown, and the job is marked as
		// not attempted with the correct status, this is retryable.
		runnerStatus, err := w.getRunnerStatus(ctx, job.RunnerID)
		if err != nil {
			return false, err
		}
		if runnerStatus != app.RunnerStatusActive {
			l.Error("runner marked unhealthy during job")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusUnknown, "runner became unhealthy during job")
			w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusUnknown)
			return true, nil
		}

		// check the runner to make sure it did not become unhealthy from the time it picked up the execution,
		// to the current time
		executionStatus, err := activities.AwaitGetJobExecutionStatus(ctx, activities.GetJobExecutionStatusRequest{
			JobExecutionID: jobExecution.ID,
		})
		if err != nil {
			return false, err
		}

		// handle the job execution status
		switch executionStatus {
		case app.RunnerJobExecutionStatusFinished:
			l.Info("job execution successfully finished")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFinished, "finished")
			return false, nil
		case app.RunnerJobExecutionStatusCancelled:
			l.Info("job cancelled")
			return true, nil
		case app.RunnerJobExecutionStatusFailed:
			l.Info("job execution cancelled")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, "failed")
			return true, nil
		case app.RunnerJobExecutionStatusTimedOut:
			l.Info("job execution timed out")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, "execution timed out")
			return true, nil
		case app.RunnerJobExecutionStatusNotAttempted:
			l.Info("job execution not attempted")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, "execution not attempted")
			return true, nil
		default:
			continue
		}
	}
}
