package jobs

import (
	"context"

	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
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

// APIClientSetter lets the job loop swap in a job-token client for one execution.
// Handlers run at most one job at a time.
type APIClientSetter interface {
	SetAPIClient(nuonrunner.Client) nuonrunner.Client
}

type StatefulJobHandler interface {
	Reset(ctx context.Context) error
}
