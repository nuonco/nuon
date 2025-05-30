package jobs

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"
)

type JobHandler interface {
	Name() string

	JobType() models.AppRunnerJobType
	JobStatus() models.AppRunnerJobStatus

	// the following methods are called _in order_ in each handler
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
