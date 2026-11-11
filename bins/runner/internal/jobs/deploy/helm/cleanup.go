package helm

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	// A recovery never fetches a chart, so there is no archive to clean up.
	if h.state == nil || h.state.arch == nil {
		return nil
	}

	err := h.state.arch.Cleanup(ctx)
	h.state.arch = nil
	if err != nil {
		h.errRecorder.Record("unable to cleanup archive", err)
	}

	return err
}
