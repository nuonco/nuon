package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateOrUpdateSkippedInstallGroupRunInput struct {
	AppBranchRunID   string `json:"app_branch_run_id" validate:"required"`
	InstallGroupID   string `json:"install_group_id" validate:"required"`
	InstallGroupName string `json:"install_group_name" validate:"required"`
	TotalInstalls    int    `json:"total_installs"`
}

// CreateOrUpdateSkippedInstallGroupRun records that a group's plan step was
// user-skipped, so the branch UI can show the group run as skipped instead of
// missing entirely. If the deploy step already created a run for this
// run+group (e.g. a retry after a prior partial deploy), that row is updated
// in place rather than duplicated.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateOrUpdateSkippedInstallGroupRun(ctx context.Context, input *CreateOrUpdateSkippedInstallGroupRunInput) error {
	now := time.Now()
	status := app.CompositeStatus{
		Status:                 app.StatusUserSkipped,
		StatusHumanDescription: "install group plan skipped",
	}

	var existing app.InstallGroupRun
	err := a.db.WithContext(ctx).
		Where(app.InstallGroupRun{
			AppBranchRunID: input.AppBranchRunID,
			InstallGroupID: input.InstallGroupID,
		}).
		Order("created_at DESC").
		First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("unable to get install group run: %w", err)
	}

	if err == nil {
		res := a.db.WithContext(ctx).
			Model(&app.InstallGroupRun{ID: existing.ID}).
			Updates(map[string]any{
				"status":       status,
				"completed_at": &now,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update install group run: %w", res.Error)
		}
		return nil
	}

	run := app.InstallGroupRun{
		AppBranchRunID:   input.AppBranchRunID,
		InstallGroupID:   input.InstallGroupID,
		InstallGroupName: input.InstallGroupName,
		TotalInstalls:    input.TotalInstalls,
		StartedAt:        &now,
		CompletedAt:      &now,
		Status:           status,
	}
	if err := a.db.WithContext(ctx).Create(&run).Error; err != nil {
		return fmt.Errorf("unable to create install group run: %w", err)
	}

	return nil
}
