package jobloop

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
)

func (j *jobLoop) executeExecuteJobStep(ctx context.Context, handler jobs.JobHandler, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if j.settings.SandboxMode {
		j.l.Info("sandbox mode enabled, skipping", zap.String("step", "execute"))
		return nil
	}

	if err := handler.Exec(ctx, job, jobExecution); err != nil {
		return fmt.Errorf("unable to execute job: %w", err)
	}

	return nil
}
