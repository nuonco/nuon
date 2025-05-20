package helm

import (
	"context"
	"errors"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if h.state.cfg == nil {
		return errors.New("no helm build config")
	}

	return nil
}
