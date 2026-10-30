package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildselection"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildsource"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
)

type FindReusableSandboxBuildInput struct {
	AppID       string `json:"app_id" validate:"required"`
	AppConfigID string `json:"app_config_id" validate:"required"`
	RunID       string `json:"run_id" validate:"required"`
}

type FindReusableSandboxBuildOutput struct {
	BuildID        string                   `json:"build_id,omitempty"`
	SourceIdentity *sourceidentity.Identity `json:"source_identity,omitempty"`
}

func (a *Activities) FindReusableSandboxBuild(ctx context.Context, input *FindReusableSandboxBuildInput) (*FindReusableSandboxBuildOutput, error) {
	out := &FindReusableSandboxBuildOutput{}
	if err := a.v.Struct(input); err != nil {
		return out, nil
	}

	cfg, err := a.getAppSandboxConfigByAppConfigID(ctx, input.AppConfigID)
	if err != nil {
		return out, nil
	}

	return a.findReusableSandboxBuild(ctx, input.AppID, input.RunID, cfg)
}

func (a *Activities) findReusableSandboxBuild(ctx context.Context, appID, runID string, cfg *app.AppSandboxConfig) (*FindReusableSandboxBuildOutput, error) {
	out := &FindReusableSandboxBuildOutput{}
	run, err := a.getAppBranchRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run.Force {
		return out, nil
	}
	var appConfig app.AppConfig
	if err := a.db.WithContext(ctx).Where(app.AppConfig{ID: cfg.AppConfigID}).First(&appConfig).Error; err != nil {
		return nil, err
	}
	var build *app.AppSandboxBuild
	if appConfig.BuildsPinned {
		if run.Comparison == nil || run.Comparison.BaseRun == nil || run.Comparison.BaseRun.AppConfigID == "" {
			return out, nil
		}
		build, err = configdiff.LatestActiveSandboxBuild(ctx, a.db, run.Comparison.BaseRun.AppConfigID)
		if err != nil {
			return nil, err
		}
		if build == nil {
			return out, nil
		}
	} else {
		build, err = a.latestEquivalentSandboxBuild(ctx, appID, cfg)
		if err != nil {
			return nil, err
		}
	}
	if build == nil || buildsource.SandboxChecksum(cfg) != buildsource.SandboxChecksum(&build.AppSandboxConfig) {
		return out, nil
	}
	identity, err := buildsource.SandboxSelection(ctx, a.db, a.vcsHelpers, &app.AppSandboxBuild{AppSandboxConfig: *cfg, AppBranchRunID: &runID})
	if err != nil {
		return nil, err
	}
	equal, err := buildsource.SourcesEqual(ctx, a.vcsHelpers, cfg.ConnectedGithubVCSConfig, cfg.PublicGitVCSConfig, build.SourceIdentity, identity)
	if err != nil {
		return nil, err
	}
	if !equal {
		return out, nil
	}
	out.BuildID = build.ID
	return out, buildselection.PinSandbox(ctx, a.db, cfg.AppConfigID, build.ID)
}

func (a *Activities) latestEquivalentSandboxBuild(ctx context.Context, appID string, cfg *app.AppSandboxConfig) (*app.AppSandboxBuild, error) {
	var build app.AppSandboxBuild
	err := a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Where(app.AppSandboxBuild{
			AppID:  appID,
			Status: app.AppSandboxBuildStatusActive,
		}).
		Order("created_at DESC").
		First(&build).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get latest sandbox build: %w", err)
	}
	if sandboxConfigsEquivalent(cfg, &build.AppSandboxConfig) {
		return &build, nil
	}
	return nil, nil
}

func sandboxConfigsEquivalent(a, b *app.AppSandboxConfig) bool {
	if a == nil || b == nil {
		return false
	}
	aRepo, _ := sandboxConfigRepo(a)
	bRepo, _ := sandboxConfigRepo(b)
	if !repoURLsEqual(aRepo, bRepo) || sandboxConfigBranch(a) != sandboxConfigBranch(b) {
		return false
	}
	if normalizeRepoPath(sandboxConfigDirectory(a)) != normalizeRepoPath(sandboxConfigDirectory(b)) {
		return false
	}
	return buildsource.SandboxChecksum(a) == buildsource.SandboxChecksum(b)
}

func sandboxConfigDirectory(cfg *app.AppSandboxConfig) string {
	if cfg == nil {
		return ""
	}
	if cfg.ConnectedGithubVCSConfig != nil {
		return cfg.ConnectedGithubVCSConfig.Directory
	}
	if cfg.PublicGitVCSConfig != nil {
		return cfg.PublicGitVCSConfig.Directory
	}
	return ""
}

func AwaitFindReusableSandboxBuild(ctx workflow.Context, input *FindReusableSandboxBuildInput, opts ...*workflow.ActivityOptions) (*FindReusableSandboxBuildOutput, error) {
	var result *FindReusableSandboxBuildOutput
	options := workflow.GetActivityOptions(ctx)
	options.StartToCloseTimeout = 5 * time.Minute
	options.RetryPolicy = &temporal.RetryPolicy{MaximumAttempts: 3}
	for _, opt := range opts {
		if opt != nil && opt.StartToCloseTimeout != 0 {
			options.StartToCloseTimeout = opt.StartToCloseTimeout
		}
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	err := workflow.ExecuteActivity(ctx, (*Activities).FindReusableSandboxBuild, input).Get(ctx, &result)
	return result, err
}
