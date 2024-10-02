package jobloop

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/sourcegraph/conc/panics"
)

func (j *jobLoop) updateJobExecutionStatus(ctx context.Context, jobID, jobExecutionID string, status models.AppRunnerJobExecutionStatus) error {
	if _, err := j.apiClient.UpdateJobExecution(ctx, jobID, jobExecutionID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: status,
	}); err != nil {
		return fmt.Errorf("unable to update job execution status: %w", err)
	}

	return nil
}

func (j *jobLoop) errToStatus(err error) models.AppRunnerJobExecutionStatus {
	if errors.Is(err, context.DeadlineExceeded) {
		return models.AppRunnerJobExecutionStatusTimedDashOut
	}

	return models.AppRunnerJobExecutionStatusFailed
}

func (j *jobLoop) execJobStep(ctx context.Context, step *executeJobStep, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if err := j.updateJobExecutionStatus(ctx, job.ID, jobExecution.ID, step.startStatus); err != nil {
		return err
	}

	var (
		pc  panics.Catcher
		err error
	)
	pc.Try(func() {
		err = step.fn(ctx, step.handler, job, jobExecution)
	})
	// when a job handler panics, we update the job to a failed status, and propagate the error
	recovered := pc.Recovered()
	if recovered != nil {
		status := models.AppRunnerJobExecutionStatusFailed
		if updateErr := j.updateJobExecutionStatus(ctx, job.ID, jobExecution.ID, status); updateErr != nil {
			j.errRecorder.Record("update_job_execution", updateErr)
		}

		// TODO(sdboyer) panics need their own error handling, can't have a handler able to blow up the whole runner
		// panic(recovered)
	}

	if err == nil {
		return nil
	}

	status := j.errToStatus(err)
	if updateErr := j.updateJobExecutionStatus(ctx, job.ID, jobExecution.ID, status); updateErr != nil {
		j.errRecorder.Record("update_job_execution", updateErr)
	}

	if step.cleanupFn == nil {
		return err
	}
	if cleanupErr := step.cleanupFn(ctx, step.handler, job, jobExecution); cleanupErr != nil {
		j.errRecorder.Record("cleanup", cleanupErr)
	}

	return err
}
