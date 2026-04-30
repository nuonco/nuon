package activities

import (
	"context"
	"fmt"
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
	res := a.db.WithContext(ctx).Exec(`
		UPDATE runner_processes
		SET composite_status = jsonb_set(
			jsonb_set(
				COALESCE(composite_status::jsonb, '{}'::jsonb),
				'{metadata}',
				COALESCE(composite_status::jsonb -> 'metadata', '{}'::jsonb)
			),
			'{metadata,shutdown_requested}',
			'true'::jsonb
		)
		WHERE deleted_at = 0
		AND composite_status::jsonb ->> 'status' IN ('active', 'offline')
	`)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to mark active processes for shutdown: %w", res.Error)
	}

	return &MarkActiveProcessesForShutdownResponse{RowsAffected: res.RowsAffected}, nil
}
