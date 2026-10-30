package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildsource"
)

type CheckSandboxBuildNeededInput struct {
	NewAppConfigID string `json:"new_app_config_id"`
	OldAppConfigID string `json:"old_app_config_id"`
	RunID          string `json:"run_id"`
	Force          bool   `json:"force"`
}

type CheckSandboxBuildNeededOutput struct {
	NeedsBuild      bool                     `json:"needs_build"`
	ExistingBuildID string                   `json:"existing_build_id,omitempty"`
	ChangeReason    string                   `json:"change_reason,omitempty"`
	SourceIdentity  *sourceidentity.Identity `json:"source_identity,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) CheckSandboxBuildNeeded(ctx context.Context, input *CheckSandboxBuildNeededInput) (*CheckSandboxBuildNeededOutput, error) {
	if input.Force || input.OldAppConfigID == "" {
		return &CheckSandboxBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
	}

	newCfg, err := a.getAppSandboxConfigByAppConfigID(ctx, input.NewAppConfigID)
	if err != nil {
		return &CheckSandboxBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
	}

	oldCfg, err := a.getAppSandboxConfigByAppConfigID(ctx, input.OldAppConfigID)
	if err != nil {
		return &CheckSandboxBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
	}

	if buildsource.SandboxChecksum(oldCfg) != buildsource.SandboxChecksum(newCfg) {
		return &CheckSandboxBuildNeededOutput{
			NeedsBuild:   true,
			ChangeReason: ChangeReasonConfigChanged,
		}, nil
	}

	if input.RunID == "" {
		return &CheckSandboxBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
	}

	reused, reuseErr := a.findReusableSandboxBuild(ctx, newCfg.AppID, input.RunID, newCfg)
	if reuseErr != nil {
		return nil, fmt.Errorf("unable to check sandbox source reuse: %w", reuseErr)
	}
	if reused != nil && reused.BuildID != "" {
		return reuseExistingSandboxBuild(reused.BuildID), nil
	}

	return &CheckSandboxBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged, SourceIdentity: reused.SourceIdentity}, nil
}

func reuseExistingSandboxBuild(existingBuildID string) *CheckSandboxBuildNeededOutput {
	return &CheckSandboxBuildNeededOutput{
		NeedsBuild:      false,
		ExistingBuildID: existingBuildID,
		ChangeReason:    ChangeReasonNoChanges,
	}
}
