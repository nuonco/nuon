package sandboxctl

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

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

func (s *Server) syncFromAPI(ctx context.Context) {
	configs, err := s.apiClient.GetSandboxConfigs(ctx)
	if err != nil {
		s.l.Warn("unable to sync sandbox configs from API", zap.Error(err))
		return
	}
	if len(configs) > 0 {
		s.state.SyncFromAPI(configs)
		s.l.Debug("synced sandbox configs from API", zap.Int("count", len(configs)))
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

			// Periodic sync from API every 30 seconds
			s.wg.Go(func() {
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-bgCtx.Done():
						return
					case <-ticker.C:
						s.syncFromAPI(bgCtx)
					}
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
