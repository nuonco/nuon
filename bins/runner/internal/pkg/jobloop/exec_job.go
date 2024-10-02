package jobloop

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
)

type executeJobStep struct {
	name      string
	fn        func(context.Context, jobs.JobHandler, *models.AppRunnerJob, *models.AppRunnerJobExecution) error
	cleanupFn func(context.Context, jobs.JobHandler, *models.AppRunnerJob, *models.AppRunnerJobExecution) error
	handler   jobs.JobHandler

	startStatus models.AppRunnerJobExecutionStatus
}

func (j *jobLoop) executeJob(ctx context.Context, job *models.AppRunnerJob) error {
	// create an execution in the API
	execution, err := j.apiClient.CreateJobExecution(ctx, job.ID, new(models.ServiceCreateRunnerJobExecutionRequest))
	if err != nil {
		return fmt.Errorf("unable to create execution: %w", err)
	}

	handler, err := j.getHandler(job)
	if err != nil {
		if err := j.updateJobExecutionStatus(ctx, job.ID, execution.ID, models.AppRunnerJobExecutionStatusFailed); err != nil {
			j.errRecorder.Record("no handler found", err)
		}

		return fmt.Errorf("unable to get handler for job: %w", err)
	}

	j.l.Info("handling job", zap.String("name", handler.Name()))
	steps := []*executeJobStep{
		// validate step
		{
			name:        "fetching",
			fn:          j.executeFetchJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInitializing,
		},
		// validate step
		{
			name:        "validate",
			fn:          j.executeValidateJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInitializing,
		},
		// initialize step
		{
			name:        "initialize",
			fn:          j.executeInitializeJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInitializing,
		},
		// execute step
		{
			name:        "execute",
			fn:          j.executeExecuteJobStep,
			cleanupFn:   j.cleanupJobStep,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInDashProgress,
		},
		// update clean up
		{
			name:        "cleanup",
			fn:          j.executeCleanupJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusCleaningDashUp,
		},
	}

	for _, step := range steps {
		j.l.Info("executing job step", zap.String("step", step.name))
		if err := j.execJobStep(ctx, step, job, execution); err != nil {
			return err
		}
	}

	if err := j.updateJobExecutionStatus(ctx, job.ID, execution.ID, models.AppRunnerJobExecutionStatusFinished); err != nil {
		return fmt.Errorf("unable to update job execution status after successful execution: %w", err)
	}

	j.l.Info("finished job", zap.String("name", handler.Name()))

	return nil
}
