// Package health serves a minimal /healthz endpoint from the mng process.
//
// Azure VMSS automatic instance repair needs an application-level health
// signal. The VMSS Application Health extension probes this endpoint; while the
// mng process is up it returns 200, and when the runner powers the VM off (see
// process.ShutdownPoller) the probe fails, the instance is marked unhealthy,
// and the VMSS repairs it. This mirrors the AWS ASG EC2 health check, which is
// what brings a shut-down AWS runner back automatically.
package health

import (
	"context"
	"fmt"
	"net/http"
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
	l   *zap.Logger
	srv *http.Server
	wg  *conc.WaitGroup
}

func New(params Params) (*Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/livez", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	s := &Server{
		l: params.L,
		srv: &http.Server{
			// Bind to loopback only: the Azure VMSS Application Health
			// extension runs inside the guest and probes 127.0.0.1, so the
			// endpoint never needs to be reachable off-host.
			Addr:              fmt.Sprintf("127.0.0.1:%d", params.Cfg.HealthPort),
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		wg: conc.NewWaitGroup(),
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
