package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ReconcileInstallBranchForLabelsInput struct {
	InstallID string `json:"install_id" validate:"required"`
}

type ReconcileInstallBranchForLabelsOutput struct {
	AppBranchID    string `json:"app_branch_id,omitempty"`
	InstallGroupID string `json:"install_group_id,omitempty"`
}

// @temporal-gen-v2 activity
func (a *Activities) ReconcileInstallBranchForLabels(ctx context.Context, input *ReconcileInstallBranchForLabelsInput) (*ReconcileInstallBranchForLabelsOutput, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).Where(app.Install{ID: input.InstallID}).First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	matches, err := a.appsHelpers.FindBranchesMatchingLabels(ctx, install.AppID, install.Labels)
	if err != nil {
		return nil, err
	}
	if len(matches) != 1 {
		return &ReconcileInstallBranchForLabelsOutput{}, nil
	}

	match := matches[0]
	a.appsHelpers.SyncInstallBranchConnection(ctx, &install, match.Branch.ID)
	return &ReconcileInstallBranchForLabelsOutput{
		AppBranchID:    match.Branch.ID,
		InstallGroupID: match.Group.ID,
	}, nil
}
