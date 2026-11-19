package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

type GetLatestActiveBranchAppConfigInput struct {
	AppBranchID string `json:"app_branch_id" validate:"required"`
	InstallID   string `json:"install_id" validate:"required"`
}

type GetLatestActiveBranchAppConfigOutput struct {
	AppConfigID    string `json:"app_config_id,omitempty"`
	AlreadyCurrent bool   `json:"already_current"`
	DesiredOnly    bool   `json:"desired_only"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetLatestActiveBranchAppConfig(ctx context.Context, input *GetLatestActiveBranchAppConfigInput) (*GetLatestActiveBranchAppConfigOutput, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).First(&install, "id = ?", input.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	run, err := a.installHelpers.LatestDeployableAppBranchRun(ctx, input.AppBranchID)
	if errors.Is(err, installhelpers.ErrNoDeployableAppBranchRun) {
		return &GetLatestActiveBranchAppConfigOutput{}, nil
	}
	if err != nil {
		return nil, err
	}

	out := &GetLatestActiveBranchAppConfigOutput{
		AppConfigID:    run.AppConfigID,
		AlreadyCurrent: install.DeployedAppConfigID() == run.AppConfigID,
	}
	if out.AlreadyCurrent {
		return out, nil
	}

	enabled, err := a.features.OrgHasFeature(ctx, install.OrgID, app.OrgFeatureDesiredConfigRollout)
	if err != nil {
		return nil, fmt.Errorf("unable to check desired config rollout feature: %w", err)
	}
	if enabled {
		if install.AppConfigID == run.AppConfigID {
			out.AlreadyCurrent = true
			return out, nil
		}
		out.DesiredOnly = true
	}
	return out, nil
}
