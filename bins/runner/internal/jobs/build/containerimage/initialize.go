package containerimage

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	// NOOP
	return nil
}
