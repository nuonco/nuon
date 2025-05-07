package terraform

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if err := h.state.arch.Cleanup(ctx); err != nil {
		h.errRecorder.Record("unable to cleanup archive", err)
	}

	if err := h.state.tfWorkspace.Cleanup(ctx); err != nil {
		return errors.Wrap(err, "unable to cleanup")
	}

	return nil
}
