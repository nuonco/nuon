package jobloop

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegraph/conc/pool"
	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/audit"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/drain"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	"github.com/nuonco/nuon/pkg/metrics"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

type JobLoop interface {
	Start() error
	Stop() error
	LifecycleHook() fx.Hook
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
	cfg      *runnerconfig.Config

	pollCtx    context.Context
	pollCancel func()
	jobCtx     context.Context
	jobCancel  func()

	l          *zap.Logger
	mw         metrics.Writer
	shutdowner fx.Shutdowner

	processRegistrar *process.Registrar
	audit            *audit.Writer

	drainer   *drain.Drainer
	jobDoneCh chan struct{}

	coalescersMu sync.Mutex
	coalescers   map[string]*statusCoalescer

	idleFn   func(context.Context)
	lastIdle time.Time

	healthcheck Healthcheck
}

type Option func(*jobLoop)

func WithIdleHook(fn func(context.Context)) Option {
	return func(j *jobLoop) {
		j.idleFn = fn
	}
}

func New(handlers []jobs.JobHandler, jobGroup models.AppRunnerJobGroup, params BaseParams, opts ...Option) *jobLoop {
	pollCtx, pollCancel := context.WithCancel(context.Background())
	jobCtx, jobCancel := context.WithCancel(context.Background())

	jobDoneCh := make(chan struct{})
	params.Drainer.Register(jobDoneCh)

	jl := &jobLoop{
		apiClient:   params.Client,
		errRecorder: params.ErrRecorder,

		jobGroup:    jobGroup,
		jobHandlers: handlers,

		pool:       pool.New().WithMaxGoroutines(1),
		pollCtx:    pollCtx,
		pollCancel: pollCancel,
		jobCtx:     jobCtx,
		jobCancel:  jobCancel,
		l:          params.L,
		settings:   params.Settings,
		cfg:        params.Cfg,
		mw:         params.MW,
		shutdowner: params.Shutdowner,

		processRegistrar: params.ProcessRegistrar,
		audit:            params.Audit,

		drainer:    params.Drainer,
		jobDoneCh:  jobDoneCh,
		coalescers: make(map[string]*statusCoalescer),
	}

	for _, opt := range opts {
		opt(jl)
	}

	params.LC.Append(jl.LifecycleHook())

	return jl
}
