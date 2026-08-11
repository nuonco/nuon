package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
)

type ComputeInstallConfigDiffInput struct {
	OldAppConfigID string `json:"old_app_config_id"`
	NewAppConfigID string `json:"new_app_config_id" validate:"required"`
}

type ComputeInstallConfigDiffOutput struct {
	Diff *app.InstallConfigDiff `json:"diff"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ComputeInstallConfigDiff(ctx context.Context, input *ComputeInstallConfigDiffInput) (*ComputeInstallConfigDiffOutput, error) {
	diff, err := configdiff.ComputeInstallConfigDiff(ctx, a.db, input.OldAppConfigID, input.NewAppConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to compute install config diff: %w", err)
	}
	return &ComputeInstallConfigDiffOutput{Diff: diff}, nil
}
