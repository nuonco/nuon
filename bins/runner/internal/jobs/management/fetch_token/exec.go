package fetchtoken

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
)

func (h *handler) finishJob(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	_, err := h.apiClient.UpdateJobExecution(ctx, job.ID, jobExecution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: models.AppRunnerJobExecutionStatusFinished,
	})
	if err != nil {
		return err
	}

	if _, err := h.apiClient.UpdateJob(ctx, job.ID, &models.ServiceUpdateRunnerJobRequest{
		Status: models.AppRunnerJobStatusFinished,
	}); err != nil {
		return err
	}
	return nil
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("exec", zap.String("job_type", "fetch_token"))

	l.Info("authenticating with AWS presigned requests")
	result, err := FetchAndStoreToken(ctx, h.apiClient)
	if err != nil {
		return err
	}

	l.Info("authentication successful",
		zap.String("runner_id", result.RunnerID),
		zap.String("instance_id", result.InstanceID),
		zap.String("aws_account_id", result.AccountID))

	l.Info("token written successfully", zap.String("path", result.TokenPath))

	if err := h.finishJob(ctx, job, jobExecution); err != nil {
		return fmt.Errorf("failed to finish job: %w", err)
	}

	return nil
}
