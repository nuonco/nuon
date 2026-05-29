package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ActiveProcessRef struct {
	RunnerID  string `json:"runner_id"`
	ProcessID string `json:"process_id"`
}

type ListActiveProcessesForHealthCheckRequest struct{}

// ListActiveProcessesForHealthCheck returns the active/offline runner processes —
// the same set the process_healthcheck signal acts on (others are noop). The
// status filter is applied in SQL via the composite_status JSONB column.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ListActiveProcessesForHealthCheck(ctx context.Context, _ *ListActiveProcessesForHealthCheckRequest) ([]ActiveProcessRef, error) {
	var processes []app.RunnerProcess
	if res := a.db.WithContext(ctx).
		Select("id", "runner_id").
		Where("composite_status->>'status' IN ?", []string{
			string(app.RunnerProcessStatusActive),
			string(app.RunnerProcessStatusOffline),
		}).
		Find(&processes); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to list active processes")
	}

	refs := make([]ActiveProcessRef, 0, len(processes))
	for _, p := range processes {
		refs = append(refs, ActiveProcessRef{RunnerID: p.RunnerID, ProcessID: p.ID})
	}

	return refs, nil
}
