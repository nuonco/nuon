package process

import (
	"context"
	"os"
	"time"

	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

const (
	shutdownPollInterval = 5 * time.Second
	forceExitTimeout     = 5 * time.Second
)

type ShutdownPollerParams struct {
	fx.In

	APIClient  nuonrunner.Client
	L          *zap.Logger `name:"system"`
	LC         fx.Lifecycle
	Registrar  *Registrar
	Shutdowner fx.Shutdowner
}

type ShutdownPoller struct {
	apiClient  nuonrunner.Client
	l          *zap.Logger
	registrar  *Registrar
	shutdowner fx.Shutdowner

	ctx      context.Context
	cancelFn func()
	wg       *conc.WaitGroup
}

func NewShutdownPoller(params ShutdownPollerParams) *ShutdownPoller {
	ctx, cancelFn := context.WithCancel(context.Background())

	sp := &ShutdownPoller{
		apiClient:  params.APIClient,
		l:          params.L,
		registrar:  params.Registrar,
		shutdowner: params.Shutdowner,
		ctx:        ctx,
		cancelFn:   cancelFn,
		wg:         conc.NewWaitGroup(),
	}

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			sp.wg.Go(func() { sp.loop(sp.ctx) })
			return nil
		},
		OnStop: func(context.Context) error {
			sp.cancelFn()
			sp.wg.Wait()
			return nil
		},
	})

	return sp
}

func (sp *ShutdownPoller) loop(ctx context.Context) {
	ticker := time.NewTicker(shutdownPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		sp.check(ctx)
	}
}

func (sp *ShutdownPoller) check(ctx context.Context) {
	processID := sp.registrar.ProcessID()
	if processID == "" {
		return
	}

	proc, err := sp.apiClient.GetProcess(ctx, processID)
	if err != nil {
		sp.l.Warn("unable to poll process for shutdown", zap.Error(err))
		return
	}

	for _, shutdown := range proc.Shutdowns {
		if shutdown == nil {
			continue
		}
		if shutdown.Status == "requested" {
			sp.l.Info("shutdown requested, completing shutdown with control plane",
				zap.String("process_id", processID),
				zap.String("shutdown_id", shutdown.ID),
				zap.String("shutdown_type", string(shutdown.Type)),
			)

			if _, err := sp.apiClient.CompleteShutdown(ctx, processID, shutdown.ID); err != nil {
				sp.l.Warn("unable to mark shutdown as completed", zap.Error(err))
			} else {
				sp.l.Info("shutdown completed successfully, initiating process exit",
					zap.String("process_id", processID),
					zap.String("shutdown_id", shutdown.ID),
				)
			}

			// Force-kill the process if fx.Shutdown doesn't complete in time.
			go func() {
				time.Sleep(forceExitTimeout)
				sp.l.Warn("graceful shutdown did not complete in time, forcing exit",
					zap.Duration("timeout", forceExitTimeout),
				)
				os.Exit(1)
			}()

			sp.shutdowner.Shutdown()
			return
		}
	}
}
