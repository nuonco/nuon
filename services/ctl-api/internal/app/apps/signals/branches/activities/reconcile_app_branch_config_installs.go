package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ReconcileAppBranchConfigInstallsInput struct {
	AppBranchID       string `json:"app_branch_id" validate:"required"`
	AppBranchConfigID string `json:"app_branch_config_id" validate:"required"`
}

type ReconciledAppBranchInstall struct {
	InstallID      string `json:"install_id"`
	InstallGroupID string `json:"install_group_id"`

	// ShouldSignal is set when the install changed branches or joined an
	// explicit group in this config, and so needs to reconcile against the
	// branch's latest run.
	ShouldSignal bool `json:"should_signal"`
}

type ReconcileAppBranchConfigInstallsOutput struct {
	Installs []ReconciledAppBranchInstall `json:"installs"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) ReconcileAppBranchConfigInstalls(ctx context.Context, input *ReconcileAppBranchConfigInstallsInput) (*ReconcileAppBranchConfigInstallsOutput, error) {
	var branch app.AppBranch
	if err := a.db.WithContext(ctx).Where(app.AppBranch{ID: input.AppBranchID}).First(&branch).Error; err != nil {
		return nil, fmt.Errorf("unable to get app branch: %w", err)
	}

	var config app.AppBranchConfig
	if err := a.db.WithContext(ctx).
		Preload("InstallGroups").
		Where(app.AppBranchConfig{ID: input.AppBranchConfigID, AppBranchID: input.AppBranchID}).
		First(&config).Error; err != nil {
		return nil, fmt.Errorf("unable to get app branch config: %w", err)
	}

	previousMembers, err := a.previousExplicitInstallIDs(ctx, &config)
	if err != nil {
		return nil, err
	}

	result := &ReconcileAppBranchConfigInstallsOutput{}
	seen := make(map[string]struct{})
	for _, group := range config.InstallGroups {
		for _, installID := range group.InstallIDs {
			if _, ok := seen[installID]; ok {
				continue
			}
			seen[installID] = struct{}{}

			var install app.Install
			if err := a.db.WithContext(ctx).Where(app.Install{ID: installID}).First(&install).Error; err != nil {
				return nil, fmt.Errorf("unable to get install %s: %w", installID, err)
			}
			if install.AppID != branch.AppID {
				return nil, fmt.Errorf("install %s does not belong to app %s", installID, branch.AppID)
			}

			claimed := !install.AppBranchID.Valid || install.AppBranchID.String != branch.ID
			if claimed {
				if err := a.helpers.SetInstallAppBranch(ctx, install.ID, branch.ID); err != nil {
					return nil, fmt.Errorf("unable to move install %s to app branch: %w", install.ID, err)
				}
			}

			_, wasMember := previousMembers[install.ID]
			result.Installs = append(result.Installs, ReconciledAppBranchInstall{
				InstallID:      install.ID,
				InstallGroupID: group.ID,
				ShouldSignal:   claimed || !wasMember,
			})
		}
	}

	return result, nil
}

// previousExplicitInstallIDs returns the installs named by ID in the config
// that came before this one, so reconciliation can tell a new group member
// apart from one that was already there.
func (a *Activities) previousExplicitInstallIDs(ctx context.Context, config *app.AppBranchConfig) (map[string]struct{}, error) {
	members := map[string]struct{}{}

	var previous app.AppBranchConfig
	err := a.db.WithContext(ctx).
		Preload("InstallGroups").
		Where(app.AppBranchConfig{AppBranchID: config.AppBranchID}).
		Where("(created_at, id) < (?, ?)", config.CreatedAt, config.ID).
		Order("created_at DESC, id DESC").
		First(&previous).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return members, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get previous app branch config: %w", err)
	}

	for _, group := range previous.InstallGroups {
		for _, installID := range group.InstallIDs {
			members[installID] = struct{}{}
		}
	}
	return members, nil
}
