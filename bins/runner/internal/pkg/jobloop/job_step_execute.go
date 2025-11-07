package jobloop

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
)

func (j *jobLoop) executeExecuteJobStep(ctx context.Context, handler jobs.JobHandler, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if j.isSandbox(job) {
		if job.Type == models.AppRunnerJobTypeActionsDashWorkflow {
			return j.execActionSandboxStep(ctx, job)
		}

		return j.execSandboxStep(ctx)
	}

	j.l.Info("executing exec job", zap.String("job_id", job.ID), zap.String("job_execution_id", jobExecution.ID))
	if err := handler.Exec(ctx, job, jobExecution); err != nil {
		j.l.Error("unable to execute exec job", zap.Error(err))
		return fmt.Errorf("unable to execute job: %w", err)
	}

	return nil
}
