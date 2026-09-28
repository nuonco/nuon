package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

const (
	runnerDisabledDescription = "This install's runner is disabled, so it cannot run jobs. Re-enable it in the install stack, then try again."
	noActiveRunnerDescription = "This install's runner is not currently active. Wait for the runner to come back online or check the runner status, then try again."
)

func (h *Helpers) requireLiveInstallRunner(ctx context.Context, installID string) error {
	disabled, err := h.installRunnerDisabled(ctx, installID)
	if err != nil {
		return err
	}
	if disabled {
		return stderr.ErrConflict{
			Err:         errors.New(runnerDisabledDescription),
			Description: runnerDisabledDescription,
		}
	}

	active, err := h.installRunnerActive(ctx, installID)
	if err != nil {
		return err
	}
	if !active {
		return stderr.ErrConflict{
			Err:         errors.New(noActiveRunnerDescription),
			Description: noActiveRunnerDescription,
		}
	}
	return nil
}

func (h *Helpers) installRunnerDisabled(ctx context.Context, installID string) (bool, error) {
	groupIDs := h.db.WithContext(ctx).
		Model(&app.RunnerGroup{}).
		Select("id").
		Where(app.RunnerGroup{OwnerID: installID, OwnerType: "installs"})

	var statuses []app.RunnerStatus
	res := h.db.WithContext(ctx).
		Model(&app.Runner{}).
		Scopes(scopes.WithDisableViews).
		Where("runner_group_id IN (?)", groupIDs).
		Order("created_at DESC").
		Limit(1).
		Pluck("status", &statuses)
	if res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to get install runner status")
	}
	if len(statuses) == 0 {
		return false, nil
	}
	return statuses[0] == app.RunnerStatusDisabled, nil
}

func (h *Helpers) installRunnerActive(ctx context.Context, installID string) (bool, error) {
	groupIDs := h.db.WithContext(ctx).
		Model(&app.RunnerGroup{}).
		Select("id").
		Where(app.RunnerGroup{OwnerID: installID, OwnerType: "installs"})

	runnerIDs := h.db.WithContext(ctx).
		Model(&app.Runner{}).
		Scopes(scopes.WithDisableViews).
		Select("id").
		Where("runner_group_id IN (?)", groupIDs)

	var count int64
	res := h.db.WithContext(ctx).
		Model(&app.RunnerProcess{}).
		Where("runner_id IN (?)", runnerIDs).
		Where("composite_status->>'status' IN ?", app.ActiveRunnerProcessStatuses()).
		Count(&count)
	if res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to check active runner process")
	}
	return count > 0, nil
}
