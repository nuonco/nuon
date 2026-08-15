package jobloop

import (
	"context"
	"sync"
	"time"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/fx"
	"go.uber.org/zap"

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

	// coalescers maps execution_id -> per-execution status writer. The
	// jobLoop type is shared by every concurrent job (parallel-runner-
	// jobs feature flag) so an execution-keyed map is required to keep
	// writers isolated.
	coalescersMu sync.Mutex
	coalescers   map[string]*statusCoalescer

	// idleFn is optional maintenance work that must not overlap a job. The
	// worker is a single goroutine, so calling it from the no-work branch is
	// what guarantees nothing this loop owns is in flight.
	idleFn   func(context.Context)
	lastIdle time.Time

	// for healthcheck
	healthcheck Healthcheck
}

// Option configures a job loop beyond its handlers and group.
type Option func(*jobLoop)

// WithIdleHook registers work to run when a poll finds no jobs, rate-limited to
// idleHookInterval. The image-actions loop uses it to collect cached action
// images only while it isn't running one.
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
