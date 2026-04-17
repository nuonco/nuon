package jobloop

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/jobs"
)

func (j *jobLoop) executeValidateJobStep(ctx context.Context, handler jobs.JobHandler, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if err := handler.Validate(ctx, job, jobExecution); err != nil {
		return fmt.Errorf("unable to validate job: %w", err)
	}
	return nil
}
