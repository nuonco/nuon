package sandboxctl

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
)

// Params are the FX dependencies for the sandbox control server.
type Params struct {
	fx.In

	LC       fx.Lifecycle
	Cfg      *internal.Config
	Settings *settings.Settings
	L        *zap.Logger `name:"system"`
}

// Server is the sandbox control HTTP server. When sandbox mode is disabled,
// methods are nil-safe no-ops.
type Server struct {
	state    *State
	fixtures *FixtureRegistry
	l        *zap.Logger
	port     int

	httpServer *http.Server
	wg         *conc.WaitGroup
}

// New creates a sandbox control server. If sandbox mode is not enabled,
// returns a stub that is safe to call but does nothing.
func New(params Params) (*Server, error) {
	if !params.Settings.SandboxMode {
		params.L.Info("sandbox control server disabled (sandbox mode is off)")
		return &Server{}, nil
	}

	registry, err := NewFixtureRegistry()
	if err != nil {
		return nil, fmt.Errorf("loading sandbox fixtures: %w", err)
	}

	s := &Server{
		state:    NewState(),
		fixtures: registry,
		l:        params.L,
		port:     params.Cfg.SandboxCtlPort,
		wg:       conc.NewWaitGroup(),
	}

	params.LC.Append(s.lifecycleHook())
	return s, nil
}

func (s *Server) lifecycleHook() fx.Hook {
	return fx.Hook{
		OnStart: func(_ context.Context) error {
			return s.start()
		},
		OnStop: func(ctx context.Context) error {
			return s.stop(ctx)
		},
	}
}

func (s *Server) start() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	s.registerRoutes(router)

	addr := fmt.Sprintf(":%d", s.port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	s.l.Info("starting sandbox control server", zap.String("addr", addr))
	s.wg.Go(func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.l.Error("sandbox control server error", zap.Error(err))
		}
	})

	return nil
}

func (s *Server) stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	s.l.Info("stopping sandbox control server")
	return s.httpServer.Shutdown(ctx)
}

// Active returns true if the server is running (sandbox mode is on).
func (s *Server) Active() bool {
	return s.state != nil
}

// State returns the sandbox control state. Returns nil if not active.
func (s *Server) State() *State {
	return s.state
}

// Fixtures returns the fixture registry. Returns nil if not active.
func (s *Server) Fixtures() *FixtureRegistry {
	return s.fixtures
}
