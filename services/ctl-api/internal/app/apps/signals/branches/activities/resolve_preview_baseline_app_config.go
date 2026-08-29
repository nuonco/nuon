package activities

import (
	"context"
	"fmt"
)

type ResolvePreviewBaselineAppConfigInput struct {
	RunID       string `json:"run_id" validate:"required"`
	AppBranchID string `json:"app_branch_id" validate:"required"`
}

type ResolvePreviewBaselineAppConfigOutput struct {
	AppConfigID string `json:"app_config_id,omitempty"`
	BaseRunID   string `json:"base_run_id,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ResolvePreviewBaselineAppConfig(ctx context.Context, input *ResolvePreviewBaselineAppConfigInput) (*ResolvePreviewBaselineAppConfigOutput, error) {
	if err := a.v.Struct(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	baseline, err := a.helpers.ResolvePreviewBaselineAppConfig(ctx, input.RunID, input.AppBranchID)
	if err != nil {
		return nil, err
	}

	return &ResolvePreviewBaselineAppConfigOutput{
		AppConfigID: baseline.AppConfigID,
		BaseRunID:   baseline.BaseRunID,
	}, nil
}
