package containerimage

import (
	"context"
	"errors"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if h.state.cfg == nil {
		return errors.New("no container image pull config found on plan")
	}

	return nil
}
