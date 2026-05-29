package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// healthCheckNoopStatuses mirrors the runner healthcheck signal's isNoopHealthCheck:
// runners in these statuses are skipped by the health check.
var healthCheckNoopStatuses = []app.RunnerStatus{
	app.RunnerStatusPending,
	app.RunnerStatusProvisioning,
	app.RunnerStatusDeprovisioning,
	app.RunnerStatusReprovisioning,
	app.RunnerStatusDeprovisioned,
	app.RunnerStatusOffline,
	app.RunnerStatusAwaitingInstallStackRun,
}

type ListActiveRunnersForHealthCheckRequest struct{}

// ListActiveRunnerIDsForHealthCheck returns the IDs of runners eligible for a
// health check (i.e. NOT in a noop status). The noop filter is applied in SQL,
// preserving the per-runner skip the signal performed.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ListActiveRunnerIDsForHealthCheck(ctx context.Context, _ *ListActiveRunnersForHealthCheckRequest) ([]string, error) {
	var ids []string
	if res := a.db.WithContext(ctx).
		Model(&app.Runner{}).
		Where("status NOT IN ?", healthCheckNoopStatuses).
		Pluck("id", &ids); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to list active runners")
	}

	return ids, nil
}
