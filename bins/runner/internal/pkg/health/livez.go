package health

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
)

type Params struct {
	fx.In

	Cfg *runnerconfig.Config
	L   *zap.Logger `name:"system"`
	LC  fx.Lifecycle
}

type Server struct {
	l         *zap.Logger
	srv       *http.Server
	wg        *conc.WaitGroup
	unhealthy atomic.Bool
}

func New(params Params) (*Server, error) {
	s := &Server{
		l:  params.L,
		wg: conc.NewWaitGroup(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/livez", s.handleLivez)

	s.srv = &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", params.Cfg.HealthPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			s.l.Info("starting health server", zap.String("addr", s.srv.Addr))
			s.wg.Go(func() {
				if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					s.l.Error("health server stopped", zap.Error(err))
				}
			})
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := s.srv.Shutdown(ctx); err != nil {
				return fmt.Errorf("unable to shut down health server: %w", err)
			}
			s.wg.Wait()
			return nil
		},
	})

	return s, nil
}

func (s *Server) handleLivez(w http.ResponseWriter, _ *http.Request) {
	if s.unhealthy.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("unhealthy"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) SetUnhealthy() {
	s.unhealthy.Store(true)
	s.l.Info("health probe set to unhealthy")
}
