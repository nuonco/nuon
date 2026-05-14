package helpers

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ShutdownProcess creates a shutdown record for a runner process, updates its
// status to pending-shutdown, writes a red health check to ClickHouse, and
// enqueues the process_shutdown signal. Only the shutdown record creation is
// treated as critical — other steps are best-effort.
func (h *Helpers) ShutdownProcess(ctx context.Context, process *app.RunnerProcess, shutdownType app.RunnerProcessShutdownType) (*app.RunnerProcessShutdown, error) {
	shutdown := app.RunnerProcessShutdown{
		RunnerProcessID: process.ID,
		Type:            shutdownType,
		CompositeStatus: app.NewCompositeStatus(ctx, app.Status(app.RunnerProcessShutdownStatusRequested)),
	}
	if res := h.db.WithContext(ctx).Create(&shutdown); res.Error != nil {
		return nil, fmt.Errorf("unable to create runner process shutdown: %w", res.Error)
	}

	// Update composite status to pending-shutdown (best-effort)
	newComposite := app.NewCompositeStatus(ctx, app.Status(app.RunnerProcessStatusPendingShutdown))
	newComposite.StatusHumanDescription = "Shutdown pending"
	newComposite.History = append([]app.CompositeStatus{process.CompositeStatus}, process.CompositeStatus.History...)
	newComposite.History[0].History = nil

	if res := h.db.WithContext(ctx).
		Model(&app.RunnerProcess{ID: process.ID}).
		Updates(app.RunnerProcess{CompositeStatus: newComposite}); res.Error != nil {
		h.l.Warn("unable to update process status to pending-shutdown", zap.Error(res.Error))
	}

	// Write a red health check to ClickHouse (best-effort)
	hc := app.RunnerHealthCheck{
		RunnerID:     process.RunnerID,
		ProcessID:    process.ID,
		RunnerStatus: app.RunnerStatusError,
	}
	if res := h.chDB.WithContext(ctx).Create(&hc); res.Error != nil {
		h.l.Warn("unable to create shutdown health check", zap.Error(res.Error))
	}

	// Enqueue the process_shutdown signal (best-effort)
	if err := h.EnqueueProcessShutdown(ctx, process); err != nil {
		h.l.Warn("unable to enqueue process shutdown signal", zap.Error(err))
	}

	return &shutdown, nil
}
