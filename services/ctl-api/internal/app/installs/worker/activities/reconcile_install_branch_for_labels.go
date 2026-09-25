package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

type ReconcileInstallBranchForLabelsInput struct {
	InstallID string `json:"install_id" validate:"required"`
}

type ReconcileInstallBranchForLabelsOutput struct {
	AppBranchID    string `json:"app_branch_id,omitempty"`
	InstallGroupID string `json:"install_group_id,omitempty"`
}

// ReconcileInstallBranchForLabels finds which install group on the install's
// own branch now targets it, so the branch can bring it up to that group's app
// config. Labels move an install between groups inside its branch; they never
// move it to another branch, so this reads AppBranchID and never writes it.
//
// @temporal-gen-v2 activity
func (a *Activities) ReconcileInstallBranchForLabels(ctx context.Context, input *ReconcileInstallBranchForLabelsInput) (*ReconcileInstallBranchForLabelsOutput, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).Where(app.Install{ID: input.InstallID}).First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	branchID := install.AppBranchID.String
	if !install.AppBranchID.Valid || branchID == "" {
		return &ReconcileInstallBranchForLabelsOutput{}, nil
	}

	groups, err := a.appsHelpers.LatestConfigInstallGroups(ctx, branchID)
	if err != nil {
		return nil, err
	}

	group, err := appshelpers.ResolveInstallGroup(groups, &install)
	if err != nil {
		return nil, err
	}
	if group != nil {
		return &ReconcileInstallBranchForLabelsOutput{
			AppBranchID:    branchID,
			InstallGroupID: group.ID,
		}, nil
	}

	return &ReconcileInstallBranchForLabelsOutput{}, nil
}
