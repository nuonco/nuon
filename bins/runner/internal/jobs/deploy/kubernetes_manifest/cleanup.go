package kubernetes_manifest

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if h.state == nil || h.state.arch == nil {
		return nil
	}
	arch := h.state.arch
	h.state.arch = nil
	return arch.Cleanup(ctx)
}
