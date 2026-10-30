package helm

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if h.state == nil {
		return nil
	}

	var errs []error
	record := func(message string, err error) {
		if err == nil {
			return
		}
		errs = append(errs, fmt.Errorf("%s: %w", message, err))
		if h.errRecorder != nil {
			h.errRecorder.Record(message, err)
		}
	}
	if h.state.source != nil {
		record("unable to close chart source", h.state.source.Close())
		h.state.source = nil
	}
	h.state.output = nil
	if h.state.arch != nil {
		record("unable to cleanup archive", h.state.arch.Cleanup(ctx))
		h.state.arch = nil
	}

	if h.state.workspace != nil {
		record("unable to cleanup workspace", h.state.workspace.Cleanup(ctx))
		h.state.workspace = nil
	}

	return errors.Join(errs...)
}
