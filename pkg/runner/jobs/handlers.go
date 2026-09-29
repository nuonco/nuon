package jobs

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type JobHandler interface {
	Name() string

	JobType() models.AppRunnerJobType
	JobStatus() models.AppRunnerJobStatus

	Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error
	Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error
	Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error
	Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error
	Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error
	GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error
	Outputs(ctx context.Context) (map[string]interface{}, error)
}

type StatefulJobHandler interface {
	Reset(ctx context.Context) error
}
