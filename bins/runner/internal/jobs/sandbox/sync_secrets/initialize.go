package terraform

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/pkg/types/outputs"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	h.state.outputs = make(outputs.SyncSecretsOutput, 0)
	return nil
}
