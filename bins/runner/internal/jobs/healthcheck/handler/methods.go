package handler

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"go.uber.org/zap"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	return nil
}

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("initializing", zap.String("job_type", "healthcheck"))
	return nil
}

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	return nil
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("executing", zap.String("job_type", "healthcheck"))

	req := new(models.ServiceCreateRunnerHealthCheckRequest)
	if _, err := h.apiClient.CreateHealthCheck(ctx, req); err != nil {
		return fmt.Errorf("unable to create health check: %w", err)
	}

	return nil
}

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("cleaning up", zap.String("job_type", "healthcheck"))
	return nil
}
