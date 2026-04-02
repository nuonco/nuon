package process

import (
	"context"
	"time"

	"github.com/sourcegraph/conc"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const defaultPollInterval = 10 * time.Second

func (r *Registrar) StartPoller(ctx context.Context) {
	wg := conc.NewWaitGroup()
	wg.Go(func() {
		r.pollLoop(ctx)
	})
}

func (r *Registrar) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.checkForShutdown(ctx)
		}
	}
}

func (r *Registrar) checkForShutdown(ctx context.Context) {
	if r.processID == "" {
		return
	}

	process, err := r.apiClient.GetProcess(ctx, r.processID)
	if err != nil {
		r.l.Warn("unable to poll runner process",
			zap.String("process_id", r.processID),
			zap.Error(err),
		)
		return
	}

	// check if there are any requested shutdowns
	for _, shutdown := range process.Shutdowns {
		if shutdown.Status == "requested" {
			r.l.Info("shutdown requested, initiating graceful shutdown",
				zap.String("process_id", r.processID),
				zap.String("shutdown_type", shutdown.Type),
			)
			if r.processLog != nil {
				r.processLog.Info("shutdown requested",
					zap.String("shutdown_type", shutdown.Type),
				)
			}

			// update process status to shutting-down
			status := "shutting-down"
			_, err := r.apiClient.UpdateProcess(ctx, r.processID, &models.ServiceUpdateRunnerProcessRequest{
				Status:            &status,
				StatusDescription: "shutdown requested via API",
			})
			if err != nil {
				r.l.Warn("unable to update process status to shutting-down",
					zap.Error(err),
				)
			}

			// trigger FX graceful shutdown
			if err := r.shutdowner.Shutdown(); err != nil {
				r.l.Error("unable to trigger fx shutdown",
					zap.Error(err),
				)
			}
			return
		}
	}
}
