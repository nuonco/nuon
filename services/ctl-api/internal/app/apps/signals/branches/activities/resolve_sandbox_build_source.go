package activities

import (
	"context"
	"fmt"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ResolveSandboxBuildSourceInput struct {
	AppConfigID string `json:"app_config_id" validate:"required"`
	RunID       string `json:"run_id" validate:"required"`
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

	gitSource, err := a.GetSandboxBuildGitSource(ctx, GetSandboxBuildGitSourceRequest{
		SandboxConfigID: sandboxConfig.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get sandbox build git source: %w", err)
	}

	run, err := a.getAppBranchRunByID(ctx, input.RunID)
	if err != nil {
		return nil, err
	}

	sandboxRepo, sandboxVCSID := sandboxConfigRepo(sandboxConfig)
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
		return out, nil
	}

	if sandboxVCSID == "" || gitSource.Ref == "" {
		return out, nil
	}

	commit, fetchErr := a.FetchCommitBySHA(ctx, &FetchCommitBySHAInput{
		VcsConfigID: sandboxVCSID,
		SHA:         gitSource.Ref,
	})
	if fetchErr != nil {
		a.l.Warn("unable to fetch sandbox repo commit; using config ref")
		return out, nil
	}
	if err := a.db.WithContext(ctx).Create(commit).Error; err != nil {
		return nil, fmt.Errorf("unable to persist sandbox commit: %w", err)
	}
	gitSource.Ref = commit.SHA
	out.VCSConnectionCommitID = &commit.ID
	return out, nil
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
