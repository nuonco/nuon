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
	// Fetch fetches the job information from ctlapi and other sources
	Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error
	// Initialize ...
	Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error
	// Validate validates the input component configs etc
	Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error
	// Exec executed the actual jon based on the job type
	Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error
	Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error
	GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error
	Outputs(ctx context.Context) (map[string]interface{}, error)
}

type StatefulJobHandler interface {
	Reset(ctx context.Context) error
}
