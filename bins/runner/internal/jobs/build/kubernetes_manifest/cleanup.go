package kubernetes_manifest

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if h.state.arch != nil {
		if err := h.state.arch.Cleanup(ctx); err != nil {
			h.errRecorder.Record("unable to cleanup archive", err)
		}
	}

	if h.state.workspace != nil {
		if err := h.state.workspace.Cleanup(ctx); err != nil {
			h.errRecorder.Record("unable to cleanup workspace", err)
		}
	}

	return nil
}
