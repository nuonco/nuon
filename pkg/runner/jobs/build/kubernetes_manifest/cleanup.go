package kubernetes_manifest

import (
	"context"
	"errors"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if h.state == nil {
		return nil
	}
	h.state.source, h.state.archiveSource = nil, nil
	var err error
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
