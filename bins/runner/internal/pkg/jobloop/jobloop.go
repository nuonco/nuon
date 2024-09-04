package jobloop

import (
	"context"

	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/nuonco/nuon-runner-go/models"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
)

const (
	defaultMaxConcurrentJobs int = 1
)

type JobLoop interface {
	Start() error
	Stop() error
	LifecycleHook() fx.Hook
}

var _ JobLoop = (*jobLoop)(nil)

type jobLoop struct {
	apiClient   nuonrunner.Client
	errRecorder *errs.Recorder

	jobGroup  models.AppRunnerJobGroup
	jobStatus models.AppRunnerJobStatus

	jobHandlers []jobs.JobHandler

	maxConcurrentJobs int
	pool              *pool.Pool

	ctx context.Context
	l   *zap.Logger
}

func New(handlers []jobs.JobHandler, jobGroup models.AppRunnerJobGroup, params BaseParams) *jobLoop {
	jl := &jobLoop{
		apiClient:         params.Client,
		ctx:               params.Ctx,
		maxConcurrentJobs: defaultMaxConcurrentJobs,
		l:                 params.L,
		errRecorder:       params.ErrRecorder,
		jobGroup:          jobGroup,
		jobHandlers:       handlers,
	}

	params.LC.Append(jl.LifecycleHook())

	return nil
}
