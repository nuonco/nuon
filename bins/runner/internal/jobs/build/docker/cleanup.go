package docker

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExec *models.AppRunnerJobExecution) error {
	if err := h.state.workspace.Cleanup(ctx); err != nil {
		h.errRecorder.Record("unable to cleanup", err)
	}
	h.state = nil

	// TODO(jm): remove the docker image from the local registry.

	return nil
}
