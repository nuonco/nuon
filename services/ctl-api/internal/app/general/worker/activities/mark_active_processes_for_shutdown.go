package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type MarkActiveProcessesForShutdownRequest struct{}

type MarkActiveProcessesForShutdownResponse struct {
	RowsAffected int64 `json:"rows_affected"`
}

// MarkActiveProcessesForShutdown sets shutdown_requested in the composite_status
// metadata of all active/offline runner processes in a single query. The
// per-process health check cron will pick this up and create shutdown jobs.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) MarkActiveProcessesForShutdown(ctx context.Context, req MarkActiveProcessesForShutdownRequest) (*MarkActiveProcessesForShutdownResponse, error) {
	query := a.db.WithContext(ctx).
		Model(&app.RunnerProcess{}).
		Scopes(generics.WhereJSONBStatusIn(
			"composite_status",
			string(app.RunnerProcessStatusActive),
			string(app.RunnerProcessStatusOffline),
		))
	res := generics.SetJSONBMetadataKey(query, "composite_status", "shutdown_requested", true)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to mark active processes for shutdown: %w", res.Error)
	}

	return &MarkActiveProcessesForShutdownResponse{RowsAffected: res.RowsAffected}, nil
}
