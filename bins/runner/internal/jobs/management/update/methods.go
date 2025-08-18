package update

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	// Ask the API what version this runner should be running
	settings, err := h.apiClient.GetSettings(ctx)

	h.state = &handlerState{}

	if err != nil {
		return errors.Wrap(err, "unable to get settings")
	}
	h.state.expectedVersion = settings.ContainerImageTag
	return nil
}

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("initializing", zap.String("job_type", "update-version"), zap.String("expected_version", h.state.expectedVersion))
	return nil
}

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("validating", zap.String("job_type", "update-version"), zap.String("expected_version", h.state.expectedVersion))
	if err := jobs.Matches(job, h); err != nil {
		return err
	}
	return nil
}

func (h *handler) Outputs(ctx context.Context) (map[string]any, error) {
	return map[string]any{}, nil
}
