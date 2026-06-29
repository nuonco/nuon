// Package health serves a minimal /livez endpoint from the mng process.
//
// Azure VMSS automatic instance repair needs an application-level health
// signal. The VMSS Application Health extension probes this endpoint; while the
// mng process is up it returns 200 and the instance is Healthy.
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

	"github.com/nuonco/nuon/bins/runner/internal"
)

type Params struct {
	fx.In

	Cfg *internal.Config
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
		// Bind to loopback only: the Azure VMSS Application Health extension
		// runs inside the vm and probes 127.0.0.1
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

// SetUnhealthy flips the /livez probe to fail (503) so the Azure VMSS
// Application Health extension marks the instance Unhealthy and automatic
// instance repair replaces it.
func (s *Server) SetUnhealthy() {
	s.unhealthy.Store(true)
	s.l.Info("health probe set to unhealthy")
}
