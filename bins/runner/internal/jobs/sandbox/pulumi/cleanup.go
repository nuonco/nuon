package pulumi

import (
	"context"
	"os"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if h.state == nil {
		return nil
	}

	if h.state.srcWorkspace != nil {
		if err := h.state.srcWorkspace.Cleanup(ctx); err != nil {
			h.errRecorder.Record("unable to cleanup source workspace", err)
		}
	}

	if h.state.workspace != nil {
		stateDir := h.state.workspace.StateDir()
		if err := os.RemoveAll(stateDir); err != nil {
			h.errRecorder.Record("unable to cleanup pulumi state directory", err)
		}
	}

	h.state = nil
	return nil
}
