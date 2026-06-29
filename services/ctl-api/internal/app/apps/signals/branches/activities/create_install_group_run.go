package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateInstallGroupRunInput struct {
	AppBranchRunID   string `json:"app_branch_run_id" validate:"required"`
	InstallGroupID   string `json:"install_group_id" validate:"required"`
	InstallGroupName string `json:"install_group_name" validate:"required"`
	TotalInstalls    int    `json:"total_installs"`
}

type CreateInstallGroupRunOutput struct {
	InstallGroupRunID string `json:"install_group_run_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallGroupRun(ctx context.Context, input *CreateInstallGroupRunInput) (*CreateInstallGroupRunOutput, error) {
	now := time.Now()
	run := app.InstallGroupRun{
		AppBranchRunID:   input.AppBranchRunID,
		InstallGroupID:   input.InstallGroupID,
		InstallGroupName: input.InstallGroupName,
		TotalInstalls:    input.TotalInstalls,
		StartedAt:        &now,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: fmt.Sprintf("deploying to %d installs", input.TotalInstalls),
		},
	}

	if err := a.db.WithContext(ctx).Create(&run).Error; err != nil {
		return nil, fmt.Errorf("unable to create install group run: %w", err)
	}

	return &CreateInstallGroupRunOutput{
		InstallGroupRunID: run.ID,
	}, nil
}
