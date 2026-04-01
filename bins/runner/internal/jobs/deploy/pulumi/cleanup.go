package pulumi

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if h.state != nil && h.state.arch != nil {
		if err := h.state.arch.Cleanup(ctx); err != nil {
			h.errRecorder.Record("unable to cleanup archive", err)
		}
	}

	return nil
}
