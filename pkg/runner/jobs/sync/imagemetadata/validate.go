package imagemetadata

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if err := h.v.Struct(h.state.plan); err != nil {
		return errors.Wrap(err, "invalid job config")
	}

	return nil
}
