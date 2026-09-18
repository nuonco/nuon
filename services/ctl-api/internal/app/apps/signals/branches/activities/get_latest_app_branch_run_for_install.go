package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

type GetLatestAppBranchRunForInstallInput struct {
	AppBranchID string `json:"app_branch_id" validate:"required"`
	InstallID   string `json:"install_id" validate:"required"`
}

type GetLatestAppBranchRunForInstallOutput struct {
	AppBranchRunID string `json:"app_branch_run_id,omitempty"`
	AppConfigID    string `json:"app_config_id,omitempty"`
	InstallGroupID string `json:"install_group_id,omitempty"`
	AlreadyCurrent bool   `json:"already_current"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetLatestAppBranchRunForInstall(ctx context.Context, input *GetLatestAppBranchRunForInstallInput) (*GetLatestAppBranchRunForInstallOutput, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).Where(app.Install{ID: input.InstallID}).First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}
	if !install.AppBranchID.Valid || install.AppBranchID.String != input.AppBranchID {
		return &GetLatestAppBranchRunForInstallOutput{}, nil
	}

	var run app.AppBranchRun
	err := a.db.WithContext(ctx).
		Where(app.AppBranchRun{AppBranchID: input.AppBranchID}).
		Where("run_type != ?", app.AppBranchRunTypeGitPreview).
		Where("app_config_id != ''").
		Order("created_at DESC").
		First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &GetLatestAppBranchRunForInstallOutput{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get latest app branch run: %w", err)
	}

	var groups []app.AppBranchInstallGroup
	if err := a.db.WithContext(ctx).
		Where(app.AppBranchInstallGroup{AppBranchConfigID: run.AppBranchConfigID}).
		Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("unable to get install groups for app branch run: %w", err)
	}

	var installGroupID string
	for i := range groups {
		if appshelpers.InstallMatchesGroup(&groups[i], &install) {
			installGroupID = groups[i].ID
			break
		}
	}

	return &GetLatestAppBranchRunForInstallOutput{
		AppBranchRunID: run.ID,
		AppConfigID:    run.AppConfigID,
		InstallGroupID: installGroupID,
		AlreadyCurrent: install.AppConfigID == run.AppConfigID,
	}, nil
}
