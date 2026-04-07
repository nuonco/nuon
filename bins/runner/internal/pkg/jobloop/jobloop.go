package jobloop

import (
	"context"
	"time"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/jobs"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/errs"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
	"github.com/nuonco/nuon/bins/runner/internal/sandboxctl"
	"github.com/nuonco/nuon/pkg/metrics"
)

type JobLoop interface {
	Start() error
	Stop() error
	LifecycleHook() fx.Hook
	// healthcheck
	GetHealthcheck() (Healthcheck, string)
	SetLatestHealthcheckAt() error
	TimeSinceLastHealthcheck() time.Duration
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

	ctx        context.Context
	ctxCancel  func()
	l          *zap.Logger
	mw         metrics.Writer
	shutdowner fx.Shutdowner

	sandboxCtl       *sandboxctl.Server
	processRegistrar *process.Registrar

	// for healthcheck
	healthcheck Healthcheck
}

func New(handlers []jobs.JobHandler, jobGroup models.AppRunnerJobGroup, params BaseParams) *jobLoop {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	jl := &jobLoop{
		apiClient:   params.Client,
		errRecorder: params.ErrRecorder,

		jobGroup:    jobGroup,
		jobHandlers: handlers,

		pool:             pool.New().WithMaxGoroutines(1),
		ctx:              ctx,
		ctxCancel:        cancelFn,
		l:                params.L,
		settings:         params.Settings,
		cfg:              params.Cfg,
		mw:               params.MW,
		shutdowner:       params.Shutdowner,
		sandboxCtl:       params.SandboxCtl,
		processRegistrar: params.ProcessRegistrar,
	}

	params.LC.Append(jl.LifecycleHook())

	return jl
}
