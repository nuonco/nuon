package jobloop

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
)

func (j *jobLoop) executeFetchJobStep(ctx context.Context, handler jobs.JobHandler, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if j.settings.SandboxMode {
		j.execSandboxStep(ctx)
		return nil
	}

	if err := handler.Fetch(ctx, job, jobExecution); err != nil {
		return fmt.Errorf("unable to fetch job: %w", err)
	}

	return nil
}
