package process

import (
	"context"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/drain"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/health"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
	pkgshutdown "github.com/nuonco/nuon/bins/runner/internal/pkg/shutdown"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	shutdownPollInterval = 5 * time.Second
	forceExitTimeout     = 5 * time.Second
	drainTimeout         = 30 * time.Minute
)

type ShutdownPollerParams struct {
	fx.In

	APIClient  nuonrunner.Client
	Cfg        *internal.Config
	L          *zap.Logger `name:"system"`
	LC         fx.Lifecycle
	Registrar  *Registrar
	Settings   *settings.Settings
	Shutdowner fx.Shutdowner
	V          *validator.Validate
	Drainer    *drain.Drainer

	// only provided in the mng process; nil in the install/run process.
	Health *health.Server `optional:"true"`
}

type ShutdownPoller struct {
	apiClient  nuonrunner.Client
	l          *zap.Logger
	v          *validator.Validate
	registrar  *Registrar
	shutdowner fx.Shutdowner
	settings   *settings.Settings
	health     *health.Server
	drainer    *drain.Drainer

	podShutdown *podShutdown

	ctx      context.Context
	cancelFn func()
	wg       *conc.WaitGroup
}

func NewShutdownPoller(params ShutdownPollerParams) *ShutdownPoller {
	ctx, cancelFn := context.WithCancel(context.Background())

	sp := &ShutdownPoller{
		apiClient:   params.APIClient,
		l:           params.L,
		v:           params.V,
		registrar:   params.Registrar,
		shutdowner:  params.Shutdowner,
		settings:    params.Settings,
		health:      params.Health,
		drainer:     params.Drainer,
		podShutdown: newPodShutdown(params.Cfg, params.Settings, params.L),
		ctx:         ctx,
		cancelFn:    cancelFn,
		wg:          conc.NewWaitGroup(),
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

	shutdowns, err := sp.apiClient.GetProcessShutdowns(ctx, processID)
	if err != nil {
		sp.l.Warn("unable to poll process for shutdown", zap.Error(err))
		return
	}

	for _, shutdown := range shutdowns {
		if shutdown == nil {
			continue
		}

		if shutdown.Status == "requested" {
			sp.l.Info("shutdown requested, entering drain mode",
				zap.String("process_id", processID),
				zap.String("shutdown_id", shutdown.ID),
				zap.String("shutdown_type", string(shutdown.Type)),
			)

			status := "pending-shutdown"
			if _, err := sp.apiClient.UpdateProcess(ctx, processID, &models.ServiceUpdateRunnerProcessRequest{
				Status:            &status,
				StatusDescription: "draining in-flight jobs before shutdown",
			}); err != nil {
				sp.l.Warn("unable to update process status to pending-shutdown", zap.Error(err))
			}

			sp.drainer.Drain()

			if err := sp.drainer.Wait(drainTimeout); err != nil {
				sp.l.Warn("drain timeout exceeded, proceeding with shutdown",
					zap.Error(err),
					zap.Duration("timeout", drainTimeout),
				)
			} else {
				sp.l.Info("all in-flight jobs completed, proceeding with shutdown")
			}

			status = "shutting-down"
			if _, err := sp.apiClient.UpdateProcess(ctx, processID, &models.ServiceUpdateRunnerProcessRequest{
				Status:            &status,
				StatusDescription: "all jobs drained, shutting down",
			}); err != nil {
				sp.l.Warn("unable to update process status to shutting-down", zap.Error(err))
			}

			if _, err := sp.apiClient.CompleteShutdown(ctx, processID, shutdown.ID); err != nil {
				sp.l.Warn("unable to mark shutdown as completed", zap.Error(err))
			} else {
				sp.l.Info("shutdown marked as completed",
					zap.String("process_id", processID),
					zap.String("shutdown_id", shutdown.ID),
				)
			}

			if sp.podShutdown != nil {
				if err := sp.podShutdown.execute(ctx); err != nil {
					sp.l.Warn("pod shutdown failed", zap.Error(err))
				}
			}

			if sp.registrar.ProcessType() == "mng" {
				if sp.settings.Platform == "azure" {
					if sp.health != nil {
						sp.l.Info("mng process shutdown - marking vm as unhealthy; letting azure vmss replace the instance")
						sp.health.SetUnhealthy()
						return
					}
					sp.l.Error("mng process shutdown on azure but health server is nil; falling back to vm poweroff")
					if err := pkgshutdown.Shutdown(ctx, sp.l, sp.v); err != nil {
						sp.l.Warn("VM shutdown failed", zap.Error(err))
					}
				} else {
					sp.l.Info("mng process shutdown: powering off VM")
					if err := pkgshutdown.Shutdown(ctx, sp.l, sp.v); err != nil {
						sp.l.Warn("VM shutdown failed", zap.Error(err))
					}
				}
			}

			go func() {
				time.Sleep(forceExitTimeout)
				sp.l.Warn("graceful shutdown did not complete in time, forcing exit",
					zap.Duration("timeout", forceExitTimeout),
				)
				os.Exit(1)
			}()

			sp.shutdowner.Shutdown(fx.ExitCode(1))
			return
		}
	}
}
