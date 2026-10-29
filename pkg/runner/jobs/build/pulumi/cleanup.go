package pulumi

import (
	"context"
	"errors"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if h.state == nil {
		return nil
	}
	err := h.state.source.Close()
	h.state.source = nil
	if h.state.arch != nil {
		err = errors.Join(err, h.state.arch.Cleanup(ctx))
		h.state.arch = nil
	}
	if h.state.workspace != nil {
		err = errors.Join(err, h.state.workspace.Cleanup(ctx))
		h.state.workspace = nil
	}
	if err != nil && h.errRecorder != nil {
		h.errRecorder.Record("unable to cleanup build", err)
	}
	return err
}
