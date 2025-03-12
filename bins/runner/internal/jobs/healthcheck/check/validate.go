package check

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	return nil
}
