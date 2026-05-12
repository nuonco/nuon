package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type RunnerHealthMetrics struct {
	// Org runner groups
	OrgRunnerGroupsTotal          int64
	OrgRunnerGroupsMissingProcess int64

	// Install runner groups
	InstallRunnerGroupsTotal          int64
	InstallRunnerGroupsMissingProcess int64
}

type GetRunnerHealthMetricsRequest struct{}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) GetRunnerHealthMetrics(ctx context.Context, req GetRunnerHealthMetricsRequest) (*RunnerHealthMetrics, error) {
	m := &RunnerHealthMetrics{}

	// Count org runner groups and those missing an active process.
	orgStats, err := a.countRunnerGroupsMissingProcess(ctx, app.RunnerGroupTypeOrg, app.RunnerProcessTypeOrg)
	if err != nil {
		return nil, fmt.Errorf("unable to get org runner health: %w", err)
	}
	m.OrgRunnerGroupsTotal = orgStats.total
	m.OrgRunnerGroupsMissingProcess = orgStats.missingProcess

	// Count install runner groups missing an active mng process.
	mngStats, err := a.countRunnerGroupsMissingProcess(ctx, app.RunnerGroupTypeInstall, app.RunnerProcessTypeMng)
	if err != nil {
		return nil, fmt.Errorf("unable to get install mng runner health: %w", err)
	}

	// Count install runner groups missing an active install process.
	installStats, err := a.countRunnerGroupsMissingProcess(ctx, app.RunnerGroupTypeInstall, app.RunnerProcessTypeInstall)
	if err != nil {
		return nil, fmt.Errorf("unable to get install runner health: %w", err)
	}

	// An install runner group is unhealthy if it's missing either process type.
	m.InstallRunnerGroupsTotal = mngStats.total
	m.InstallRunnerGroupsMissingProcess = mngStats.missingProcess + installStats.missingProcess

	return m, nil
}

type runnerGroupStats struct {
	total          int64
	missingProcess int64
}

// countRunnerGroupsMissingProcess counts runner groups of the given type that
// have at least one active runner but no active process of the specified type.
func (a *Activities) countRunnerGroupsMissingProcess(ctx context.Context, groupType app.RunnerGroupType, processType app.RunnerProcessType) (*runnerGroupStats, error) {
	stats := &runnerGroupStats{}

	// Total runner groups of this type that have at least one active runner.
	if res := a.db.WithContext(ctx).Raw(`
		SELECT count(DISTINCT rg.id)
		FROM runner_groups rg
		JOIN runners r ON r.runner_group_id = rg.id AND r.deleted_at = 0 AND r.status = 'active'
		WHERE rg.deleted_at = 0
		AND rg.type = ?
	`, string(groupType)).Scan(&stats.total); res.Error != nil {
		return nil, fmt.Errorf("unable to count runner groups: %w", res.Error)
	}

	// Runner groups with active runners but no active process of the expected type.
	if res := a.db.WithContext(ctx).Raw(`
		SELECT count(DISTINCT rg.id)
		FROM runner_groups rg
		JOIN runners r ON r.runner_group_id = rg.id AND r.deleted_at = 0 AND r.status = 'active'
		WHERE rg.deleted_at = 0
		AND rg.type = ?
		AND NOT EXISTS (
			SELECT 1 FROM runner_processes rp
			WHERE rp.runner_id = r.id
			AND rp.deleted_at = 0
			AND rp.type = ?
			AND rp.composite_status::jsonb ->> 'status' = 'active'
		)
	`, string(groupType), string(processType)).Scan(&stats.missingProcess); res.Error != nil {
		return nil, fmt.Errorf("unable to count runner groups missing process: %w", res.Error)
	}

	return stats, nil
}
