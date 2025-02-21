package jobloop

import (
	"context"

	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/nuonco/nuon-runner-go/models"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
	"github.com/powertoolsdev/mono/pkg/metrics"
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

	pool     *pool.Pool
	settings *settings.Settings
	cfg      *internal.Config

	ctx       context.Context
	ctxCancel func()
	l         *zap.Logger
	mw        metrics.Writer
}

func New(handlers []jobs.JobHandler, jobGroup models.AppRunnerJobGroup, params BaseParams) *jobLoop {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	jl := &jobLoop{
		apiClient:   params.Client,
		errRecorder: params.ErrRecorder,

		jobGroup:    jobGroup,
		jobHandlers: handlers,

		pool:      pool.New().WithMaxGoroutines(1),
		ctx:       ctx,
		ctxCancel: cancelFn,
		l:         params.L,
		settings:  params.Settings,
		cfg:       params.Cfg,
		mw:        params.MW,
	}

	params.LC.Append(jl.LifecycleHook())

	return nil
}
