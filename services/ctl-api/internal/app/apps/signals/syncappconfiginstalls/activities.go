package syncappconfiginstalls

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ActivitiesParams struct {
	fx.In
	DB *gorm.DB `name:"psql"`
}

type Activities struct {
	db *gorm.DB
}

func NewActivities(params ActivitiesParams) *Activities {
	return &Activities{db: params.DB}
}

type GetNonBranchManagedInstallIDsInput struct {
	AppID string `json:"app_id" validate:"required"`
}

type GetNonBranchManagedInstallIDsOutput struct {
	InstallIDs []string `json:"install_ids"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) GetNonBranchManagedInstallIDs(ctx context.Context, input *GetNonBranchManagedInstallIDsInput) (*GetNonBranchManagedInstallIDsOutput, error) {
	var installs []app.Install
	if err := a.db.WithContext(ctx).
		Where(app.Install{AppID: input.AppID}).
		Where(`NOT EXISTS (
			SELECT 1
			FROM install_management_policy_versions AS current_policy
			WHERE current_policy.install_id = installs.id
				AND current_policy.version = (
					SELECT MAX(policy_version.version)
					FROM install_management_policy_versions AS policy_version
					WHERE policy_version.install_id = installs.id
				)
				AND current_policy.command_authority <> ?
		)`, app.InstallAuthorityNuon).
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to query installs: %w", err)
	}

	var branchInstallGroups []app.AppBranchInstallGroup
	if err := a.db.WithContext(ctx).
		Joins("JOIN app_branch_configs ON app_branch_configs.id = app_branch_install_groups.app_branch_config_id").
		Joins("JOIN app_branches ON app_branches.id = app_branch_configs.app_branch_id AND app_branches.deleted_at = 0").
		Where("app_branches.app_id = ?", input.AppID).
		Find(&branchInstallGroups).Error; err != nil {
		return nil, fmt.Errorf("unable to query branch install groups: %w", err)
	}

	branchManaged := make(map[string]bool)
	for _, group := range branchInstallGroups {
		for _, id := range group.InstallIDs {
			branchManaged[id] = true
		}
	}

	var result []string
	for _, install := range installs {
		if !branchManaged[install.ID] {
			result = append(result, install.ID)
		}
	}

	return &GetNonBranchManagedInstallIDsOutput{InstallIDs: result}, nil
}
