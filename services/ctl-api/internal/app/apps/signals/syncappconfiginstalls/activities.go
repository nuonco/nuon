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
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to query installs: %w", err)
	}

	var branchConnections []app.InstallAppBranchConnection
	if err := a.db.WithContext(ctx).
		Joins("JOIN app_branches ON app_branches.id = install_app_branch_connections.app_branch_id AND app_branches.deleted_at = 0").
		Where("app_branches.app_id = ?", input.AppID).
		Where(app.InstallAppBranchConnection{Active: true}).
		Find(&branchConnections).Error; err != nil {
		return nil, fmt.Errorf("unable to query install app branch connections: %w", err)
	}

	branchManaged := make(map[string]bool)
	for _, connection := range branchConnections {
		branchManaged[connection.InstallID] = true
	}

	var result []string
	for _, install := range installs {
		if !branchManaged[install.ID] {
			result = append(result, install.ID)
		}
	}

	return &GetNonBranchManagedInstallIDsOutput{InstallIDs: result}, nil
}
