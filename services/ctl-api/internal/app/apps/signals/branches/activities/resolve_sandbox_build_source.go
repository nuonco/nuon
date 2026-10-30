package activities

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/vcserrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildsource"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type GetSandboxBuildConfigInput struct {
	AppConfigID string `json:"app_config_id" validate:"required"`
}

type GetSandboxBuildConfigOutput struct {
	SandboxConfig *app.AppSandboxConfig `json:"sandbox_config,omitempty"`
	Skipped       bool                  `json:"skipped,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) GetSandboxBuildConfig(ctx context.Context, input *GetSandboxBuildConfigInput) (*GetSandboxBuildConfigOutput, error) {
	sandboxConfig, err := a.getAppSandboxConfigByAppConfigID(ctx, input.AppConfigID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &GetSandboxBuildConfigOutput{Skipped: true}, nil
		}
		return nil, err
	}
	return &GetSandboxBuildConfigOutput{SandboxConfig: sandboxConfig}, nil
}

type ResolveSandboxBuildSourceInput struct {
	AppConfigID string `json:"app_config_id" validate:"required"`
	// Empty for a directly triggered build, which has no run to pin to and
	// resolves the sandbox repo's own ref instead.
	RunID   string `json:"run_id,omitempty"`
	BuildID string `json:"build_id" validate:"required"`
}

type ResolveSandboxBuildSourceOutput struct {
	SandboxConfig         *app.AppSandboxConfig    `json:"sandbox_config,omitempty"`
	GitSource             *plantypes.GitSource     `json:"git_source,omitempty"`
	VCSConnectionCommitID *string                  `json:"vcs_connection_commit_id,omitempty"`
	SourceIdentity        *sourceidentity.Identity `json:"source_identity,omitempty"`
	Skipped               bool                     `json:"skipped,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) ResolveSandboxBuildSource(ctx context.Context, input *ResolveSandboxBuildSourceInput) (*ResolveSandboxBuildSourceOutput, error) {
	sandboxConfig, err := a.getAppSandboxConfigByAppConfigID(ctx, input.AppConfigID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &ResolveSandboxBuildSourceOutput{Skipped: true}, nil
		}
		return nil, err
	}
	build := app.AppSandboxBuild{AppSandboxConfig: *sandboxConfig}
	if input.RunID != "" {
		build.AppBranchRunID = &input.RunID
	}
	if input.BuildID != "" {
		if err := a.db.WithContext(ctx).Preload("VCSConnectionCommit").Where(app.AppSandboxBuild{ID: input.BuildID}).First(&build).Error; err != nil {
			return nil, err
		}
		build.AppSandboxConfig = *sandboxConfig
	}
	identity, err := buildsource.Sandbox(ctx, a.db, a.vcsHelpers, &build)
	if err != nil {
		return nil, err
	}
	if input.BuildID != "" {
		identity, err = buildsource.SaveSandbox(ctx, a.db, input.BuildID, identity)
		if err != nil {
			return nil, err
		}
	}
	src, err := buildsource.GitSource(ctx, a.vcsHelpers, sandboxConfig.ConnectedGithubVCSConfig, sandboxConfig.PublicGitVCSConfig, identity)
	if err != nil {
		return nil, err
	}
	return &ResolveSandboxBuildSourceOutput{SandboxConfig: sandboxConfig, GitSource: src, SourceIdentity: identity, VCSConnectionCommitID: build.VCSConnectionCommitID}, nil
}

func SandboxSourceFailureDescription(err error) string {
	if vcserrors.IsGitRefNotFound(err) {
		var appErr *temporal.ApplicationError
		if errors.As(err, &appErr) && appErr.Message() != "" {
			return appErr.Message()
		}
	}
	return "unable to resolve sandbox source"
}

func (a *Activities) failSandboxBuildSource(ctx context.Context, buildID, repo, ref string, cause error) error {
	data, buildErr := compositeerrors.New(
		&vcserrors.GitRefNotFoundError{Repo: repo, Ref: ref},
		compositeerrors.WithSource("app_sandbox_builds", buildID),
	)
	if buildErr != nil {
		return fmt.Errorf("unable to build sandbox build composite error: %w", buildErr)
	}

	if res := a.db.WithContext(ctx).
		Model(&app.AppSandboxBuild{ID: buildID}).
		Select("composite_error").
		Updates(app.AppSandboxBuild{CompositeError: data}); res.Error != nil {
		return fmt.Errorf("unable to set sandbox build composite error: %w", res.Error)
	}

	return vcserrors.NewGitRefNotFound(repo, ref, cause)
}

func (a *Activities) attachSandboxBuildCommit(ctx context.Context, buildID string, commitID *string) error {
	if buildID == "" || commitID == nil || *commitID == "" {
		return nil
	}
	if res := a.db.WithContext(ctx).
		Model(&app.AppSandboxBuild{}).
		Where(app.AppSandboxBuild{ID: buildID}).
		Update("vcs_connection_commit_id", *commitID); res.Error != nil {
		return fmt.Errorf("unable to attach commit to sandbox build: %w", res.Error)
	}
	return nil
}

func sandboxConfigRepo(cfg *app.AppSandboxConfig) (repo, vcsConfigID string) {
	if cfg == nil {
		return "", ""
	}
	if cfg.ConnectedGithubVCSConfig != nil {
		return cfg.ConnectedGithubVCSConfig.Repo, cfg.ConnectedGithubVCSConfig.ID
	}
	if cfg.PublicGitVCSConfig != nil {
		return cfg.PublicGitVCSConfig.Repo, cfg.PublicGitVCSConfig.ID
	}
	return "", ""
}

func sandboxConfigBranch(cfg *app.AppSandboxConfig) string {
	if cfg == nil {
		return ""
	}
	if cfg.ConnectedGithubVCSConfig != nil {
		return cfg.ConnectedGithubVCSConfig.Branch
	}
	if cfg.PublicGitVCSConfig != nil {
		return cfg.PublicGitVCSConfig.Branch
	}
	return ""
}
