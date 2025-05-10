package terraform

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Reset(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	h.state = nil
	return nil
}
