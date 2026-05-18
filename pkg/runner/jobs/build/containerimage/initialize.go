package containerimage

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	// NOOP
	return nil
}
