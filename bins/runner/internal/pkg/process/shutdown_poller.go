package process

import (
	"context"
	"os"
	"time"

	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/k8s"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

const (
	shutdownPollInterval = 5 * time.Second
	forceExitTimeout     = 5 * time.Second
)

type ShutdownPollerParams struct {
	fx.In

	APIClient  nuonrunner.Client
	Cfg        *internal.Config
	L          *zap.Logger `name:"system"`
	LC         fx.Lifecycle
	Registrar  *Registrar
	Shutdowner fx.Shutdowner
}

type ShutdownPoller struct {
	apiClient  nuonrunner.Client
	cfg        *internal.Config
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
		cfg:        params.Cfg,
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
			sp.l.Info("shutdown requested, marking as completed and initiating graceful shutdown",
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

			sp.deletePod(ctx)

			// Force-kill the process if fx.Shutdown doesn't complete in time.
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

func (sp *ShutdownPoller) deletePod(ctx context.Context) {
	if !sp.cfg.DeletePodOnShutdown {
		return
	}

	if sp.cfg.PodName == "" || sp.cfg.PodNamespace == "" {
		sp.l.Warn("delete_pod_on_shutdown enabled but pod_name or pod_namespace not set, skipping pod deletion")
		return
	}

	sp.l.Info("deleting own pod",
		zap.String("pod_name", sp.cfg.PodName),
		zap.String("pod_namespace", sp.cfg.PodNamespace),
	)

	clientset, _, _, err := k8s.ClientsetInCluster()
	if err != nil {
		sp.l.Warn("unable to create in-cluster k8s client for pod deletion", zap.Error(err))
		return
	}

	if err := clientset.CoreV1().Pods(sp.cfg.PodNamespace).Delete(ctx, sp.cfg.PodName, metav1.DeleteOptions{}); err != nil {
		sp.l.Warn("unable to delete own pod", zap.Error(err))
		return
	}

	sp.l.Info("successfully deleted own pod")
}
