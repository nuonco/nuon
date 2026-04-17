package sandboxctl

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

type Params struct {
	fx.In

	LC         fx.Lifecycle
	Shutdowner fx.Shutdowner
	Cfg        *internal.Config
	Settings   *settings.Settings
	L          *zap.Logger `name:"system"`
	APIClient  nuonrunner.Client
}

type Server struct {
	state      *State
	httpServer *http.Server
	shutdowner fx.Shutdowner
	l          *zap.Logger
	enabled    bool
	wg         *conc.WaitGroup
	cancelFn   func()
	apiClient  nuonrunner.Client
}

func New(params Params) *Server {
	s := &Server{
		shutdowner: params.Shutdowner,
		l:          params.L,
		wg:         conc.NewWaitGroup(),
		apiClient:  params.APIClient,
	}

	if !params.Settings.SandboxMode {
		return s
	}

	s.enabled = true
	s.state = NewState(params.Cfg.SandboxJobDuration)

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	port := params.Cfg.SandboxControlPort
	if port == 0 {
		port = 9095
	}

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	params.LC.Append(s.LifecycleHook())
	return s
}

func (s *Server) GetState() *State {
	if !s.enabled {
		return nil
	}
	return s.state
}

func (s *Server) Enabled() bool {
	return s.enabled
}

// SyncNow fetches the latest configs from the API immediately.
// Call this before each job to ensure the handler uses fresh config.
func (s *Server) SyncNow(ctx context.Context) {
	if !s.enabled {
		return
	}
	s.syncFromAPI(ctx)
}

// SyncForJob fetches the sandbox config for a specific job type and operation from the API.
func (s *Server) SyncForJob(ctx context.Context, jobType, operation string) {
	if !s.enabled {
		return
	}

	cfg, err := s.apiClient.GetSandboxConfig(ctx, jobType, operation)
	if err != nil {
		s.l.Warn("sandbox-sync: unable to fetch config for job",
			zap.String("job_type", jobType),
			zap.String("operation", operation),
			zap.Error(err),
		)
		return
	}

	if cfg == nil {
		s.l.Info("sandbox-sync: no config found for job",
			zap.String("job_type", jobType),
			zap.String("operation", operation),
		)
		return
	}

	s.l.Info("sandbox-sync: fetched config for job",
		zap.String("job_type", cfg.JobType),
		zap.String("operation", cfg.Operation),
		zap.Bool("enabled", cfg.Enabled),
		zap.Duration("duration", cfg.Duration),
	)

	s.state.SyncSingleFromAPI(cfg)
}

func (s *Server) syncFromAPI(ctx context.Context) {
	configs, err := s.apiClient.GetSandboxConfigs(ctx)
	if err != nil {
		s.l.Warn("sandbox-sync: unable to sync sandbox configs from API", zap.Error(err))
		return
	}
	s.l.Info("sandbox-sync: fetched configs from API",
		zap.Int("count", len(configs)),
	)
	if len(configs) > 0 {
		s.state.SyncFromAPI(configs)
	}
}

func (s *Server) LifecycleHook() fx.Hook {
	return fx.Hook{
		OnStart: func(ctx context.Context) error {
			if !s.enabled {
				return nil
			}

			bgCtx, cancelFn := context.WithCancel(context.Background())
			s.cancelFn = cancelFn

			// Initial sync from API
			s.syncFromAPI(bgCtx)

			s.l.Info("starting sandbox control server", zap.String("addr", s.httpServer.Addr))
			s.wg.Go(func() {
				if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatal(err)
				}
			})

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if !s.enabled {
				return nil
			}

			s.l.Info("stopping sandbox control server")
			if err := s.httpServer.Shutdown(ctx); err != nil {
				return fmt.Errorf("unable to shut down sandbox control server: %w", err)
			}

			s.cancelFn()
			s.wg.Wait()
			return nil
		},
	}
}
