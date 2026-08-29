package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type PreviewRunInput struct {
	Source           app.AppBranchRunPreviewSource
	PRNumber         *int
	GitRef           string
	HeadSHA          string
	InputAppConfigID string
	Override         *app.AppBranchPreviewOverride
}

type ResolvedPreviewRun struct {
	Preview *app.AppBranchRunPreview
}

func BranchPreviewConfigOrDefault(cfg *app.AppBranchConfig) app.AppBranchPreviewConfig {
	return branchPreviewConfigOrDefault(cfg)
}

func branchPreviewConfigOrDefault(cfg *app.AppBranchConfig) app.AppBranchPreviewConfig {
	if cfg != nil && cfg.PreviewConfig != nil {
		out := *cfg.PreviewConfig
		out.Normalize()
		return out
	}
	return app.DefaultAppBranchPreviewConfig()
}

func mergePreviewConfig(branchDefaults app.AppBranchPreviewConfig, override *app.AppBranchPreviewOverride) app.AppBranchPreviewConfig {
	resolved := branchDefaults
	if override == nil {
		return resolved
	}
	if override.Mode != nil {
		resolved.Mode = *override.Mode
	}
	if override.InstallID != nil {
		resolved.InstallID = override.InstallID
		resolved.InstallName = nil
		resolved.LabelSelector = nil
	}
	return resolved
}

func (h *Helpers) ResolvePreviewInstallID(
	ctx context.Context,
	appID string,
	cfg app.AppBranchPreviewConfig,
	overrideInstallID *string,
) (installID, installName string, err error) {
	if cfg.Mode == app.AppBranchRunPreviewModeBuildOnly {
		return "", "", nil
	}

	if overrideInstallID != nil && *overrideInstallID != "" {
		var inst app.Install
		if err := h.db.WithContext(ctx).
			Where(app.Install{AppID: appID}).
			First(&inst, "id = ?", *overrideInstallID).Error; err != nil {
			return "", "", stderr.NewInvalidRequest(fmt.Errorf("preview install %q not found", *overrideInstallID))
		}
		return inst.ID, inst.Name, nil
	}

	if cfg.InstallID != nil && *cfg.InstallID != "" {
		var inst app.Install
		if err := h.db.WithContext(ctx).
			Where(app.Install{AppID: appID}).
			First(&inst, "id = ?", *cfg.InstallID).Error; err != nil {
			return "", "", stderr.NewInvalidRequest(fmt.Errorf("preview install_id %q not found", *cfg.InstallID))
		}
		return inst.ID, inst.Name, nil
	}

	if cfg.InstallName != nil && *cfg.InstallName != "" {
		var inst app.Install
		if err := h.db.WithContext(ctx).
			Where(app.Install{AppID: appID, Name: *cfg.InstallName}).
			First(&inst).Error; err != nil {
			return "", "", stderr.NewInvalidRequest(fmt.Errorf("preview install_name %q not found", *cfg.InstallName))
		}
		return inst.ID, inst.Name, nil
	}

	if cfg.LabelSelector != nil && len(cfg.LabelSelector.MatchLabels) > 0 {
		return "", "", stderr.NewInvalidRequest(fmt.Errorf("preview install must be selected when branch preview config uses label_selector"))
	}

	return "", "", stderr.NewInvalidRequest(fmt.Errorf("preview requires an install for mode %q", cfg.Mode))
}

func (h *Helpers) ListPreviewInstallCandidates(
	ctx context.Context,
	appID string,
	_ string,
	_ app.AppBranchPreviewConfig,
) ([]app.Install, error) {
	var installs []app.Install
	if err := h.db.WithContext(ctx).
		Where(app.Install{AppID: appID}).
		Preload("AppBranch").
		Order("name ASC").
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to list preview install candidates: %w", err)
	}
	return installs, nil
}

func (h *Helpers) BuildAppBranchRunPreview(
	ctx context.Context,
	appID string,
	branchConfig *app.AppBranchConfig,
	input *PreviewRunInput,
) (*app.AppBranchRunPreview, error) {
	if input == nil {
		return nil, fmt.Errorf("preview run input is required")
	}
	if !input.Source.Valid() {
		return nil, stderr.NewInvalidRequest(fmt.Errorf("invalid preview source %q", input.Source))
	}

	branchSnapshot := branchPreviewConfigOrDefault(branchConfig)
	branchSnapshot.Normalize()

	resolved := mergePreviewConfig(branchSnapshot, input.Override)
	resolved.Normalize()
	if err := resolved.Validate(); err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}

	var overrideInstallID *string
	if input.Override != nil {
		overrideInstallID = input.Override.InstallID
	}

	installID, installName, err := h.ResolvePreviewInstallID(ctx, appID, resolved, overrideInstallID)
	if err != nil {
		return nil, err
	}

	return &app.AppBranchRunPreview{
		Source:                input.Source,
		Mode:                  resolved.Mode,
		InstallID:             installID,
		InstallName:           installName,
		GitRef:                input.GitRef,
		InputAppConfigID:      input.InputAppConfigID,
		BranchPreviewConfig:   branchSnapshot,
		OverridePreviewConfig: input.Override,
		ResolvedPreviewConfig: resolved,
	}, nil
}

func MapLegacyPlanOnlyToPreviewInput(req *CreateAppBranchRunRequest) *PreviewRunInput {
	if req.Preview != nil {
		return req.Preview
	}
	isGitPreview := req.RunType == app.AppBranchRunTypeGitPreview
	isLegacyPlanOnlyPreview := req.PlanOnly && (req.PRNumber != nil || req.HeadSHA != "")
	if !isGitPreview && !isLegacyPlanOnlyPreview {
		return nil
	}
	source := app.AppBranchRunPreviewSourceBranch
	if req.PRNumber != nil {
		source = app.AppBranchRunPreviewSourcePR
	} else if req.HeadSHA != "" {
		source = app.AppBranchRunPreviewSourceCommit
	}
	if req.AppConfigID != "" {
		source = app.AppBranchRunPreviewSourceLocal
	}
	return &PreviewRunInput{
		Source:           source,
		PRNumber:         req.PRNumber,
		HeadSHA:          req.HeadSHA,
		InputAppConfigID: req.AppConfigID,
	}
}
