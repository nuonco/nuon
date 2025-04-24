package containerimage

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if err := h.v.Struct(h.state.plan); err != nil {
		return errors.Wrap(err, "invalid job config")
	}

	return nil
}
