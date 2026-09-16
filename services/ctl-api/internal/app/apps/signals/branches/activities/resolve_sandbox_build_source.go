package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/vcserrors"
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
	RunID       string `json:"run_id" validate:"required"`
	BuildID     string `json:"build_id" validate:"required"`
}

type ResolveSandboxBuildSourceOutput struct {
	SandboxConfig         *app.AppSandboxConfig `json:"sandbox_config,omitempty"`
	GitSource             *plantypes.GitSource  `json:"git_source,omitempty"`
	VCSConnectionCommitID *string               `json:"vcs_connection_commit_id,omitempty"`
	Skipped               bool                  `json:"skipped,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ResolveSandboxBuildSource(ctx context.Context, input *ResolveSandboxBuildSourceInput) (*ResolveSandboxBuildSourceOutput, error) {
	sandboxConfig, err := a.getAppSandboxConfigByAppConfigID(ctx, input.AppConfigID)
	if err != nil {
		return &ResolveSandboxBuildSourceOutput{Skipped: true}, nil
	}

	sandboxRepo, sandboxVCSID := sandboxConfigRepo(sandboxConfig)

	gitSource, err := a.GetSandboxBuildGitSource(ctx, GetSandboxBuildGitSourceRequest{
		SandboxConfigID: sandboxConfig.ID,
	})
	if err != nil {
		if vcserrors.IsGitRefNotFound(err) {
			return nil, a.failSandboxBuildSource(ctx, input.BuildID, sandboxRepo, sandboxConfigBranch(sandboxConfig), err)
		}
		return nil, fmt.Errorf("unable to get sandbox build git source: %w", err)
	}

	run, err := a.getAppBranchRunByID(ctx, input.RunID)
	if err != nil {
		return nil, err
	}

	branchRepo := ""
	if run.AppBranchConfig.ConnectedGithubVCSConfig != nil {
		branchRepo = run.AppBranchConfig.ConnectedGithubVCSConfig.Repo
	} else if run.AppBranchConfig.PublicGitVCSConfig != nil {
		branchRepo = run.AppBranchConfig.PublicGitVCSConfig.Repo
	}

	out := &ResolveSandboxBuildSourceOutput{
		SandboxConfig: sandboxConfig,
		GitSource:     gitSource,
	}

	if repoURLsEqual(sandboxRepo, branchRepo) {
		if run.VCSConnectionCommit != nil && run.VCSConnectionCommit.SHA != "" {
			gitSource.Ref = run.VCSConnectionCommit.SHA
			out.VCSConnectionCommitID = &run.VCSConnectionCommit.ID
		} else if run.HeadSHA != "" {
			gitSource.Ref = run.HeadSHA
			if run.VCSConnectionCommitID != nil {
				out.VCSConnectionCommitID = run.VCSConnectionCommitID
			}
		}
		return out, a.attachSandboxBuildCommit(ctx, input.BuildID, out.VCSConnectionCommitID)
	}

	if sandboxVCSID == "" || gitSource.Ref == "" {
		return out, nil
	}

	commit, fetchErr := a.FetchCommitBySHA(ctx, &FetchCommitBySHAInput{
		VcsConfigID: sandboxVCSID,
		SHA:         gitSource.Ref,
	})
	if fetchErr != nil {
		if vcserrors.IsGitRefNotFound(fetchErr) {
			return nil, a.failSandboxBuildSource(ctx, input.BuildID, sandboxRepo, gitSource.Ref, fetchErr)
		}
		return nil, fmt.Errorf("unable to fetch sandbox repo commit: %w", fetchErr)
	}
	if err := a.db.WithContext(ctx).Create(commit).Error; err != nil {
		return nil, fmt.Errorf("unable to persist sandbox commit: %w", err)
	}
	gitSource.Ref = commit.SHA
	out.VCSConnectionCommitID = &commit.ID
	return out, a.attachSandboxBuildCommit(ctx, input.BuildID, out.VCSConnectionCommitID)
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
	if commitID == nil || *commitID == "" {
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
