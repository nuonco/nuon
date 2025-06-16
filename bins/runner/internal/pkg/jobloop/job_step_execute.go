package jobloop

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
)

func (j *jobLoop) executeExecuteJobStep(ctx context.Context, handler jobs.JobHandler, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if j.isSandbox(job) {
		if job.Type == models.AppRunnerJobTypeActionsDashWorkflow {
			return j.execActionSandboxStep(ctx, job)
		}

		return j.execSandboxStep(ctx)
	}

	if err := handler.Exec(ctx, job, jobExecution); err != nil {
		return fmt.Errorf("unable to execute job: %w", err)
	}

	return nil
}
