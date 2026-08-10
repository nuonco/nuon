package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetAppBranchConfigByIDInput struct {
	AppBranchConfigID string `json:"app_branch_config_id"`
}

type GetAppBranchConfigByIDOutput struct {
	PostDeployRunbookIDs []string `json:"post_deploy_runbook_ids,omitempty"`
}

// GetAppBranchConfigByID reads the branch config fields the step generator needs
// to decide the shape of the run's steps.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetAppBranchConfigByID(ctx context.Context, input *GetAppBranchConfigByIDInput) (*GetAppBranchConfigByIDOutput, error) {
	var config app.AppBranchConfig
	if err := a.db.WithContext(ctx).First(&config, "id = ?", input.AppBranchConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to get app branch config: %w", err)
	}

	return &GetAppBranchConfigByIDOutput{
		PostDeployRunbookIDs: config.PostDeployRunbookIDs,
	}, nil
}
