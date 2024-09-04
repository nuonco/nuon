package jobloop

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
)

func (j *jobLoop) executeInitializeJobStep(ctx context.Context, handler jobs.JobHandler, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if err := handler.Initialize(ctx, job, jobExecution); err != nil {
		return fmt.Errorf("unable to initialize job: %w", err)
	}

	return nil
}
