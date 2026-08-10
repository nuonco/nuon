package activities

import (
	"context"
	"fmt"
	"strings"

	githubpkg "github.com/nuonco/nuon/pkg/github/repo"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CloneInstallsRepoInput struct {
	AppInstallConfigSyncID string `json:"app_install_config_sync_id" validate:"required"`
	VCSType                string `json:"vcs_type" validate:"required"`
	VCSConnectionID        string `json:"vcs_connection_id,omitempty"`
	Repo                   string `json:"repo" validate:"required"`
	Branch                 string `json:"branch" validate:"required"`
	CommitSHA              string `json:"commit_sha,omitempty"`
}

func (a *Activities) getConnectedGitSource(ctx context.Context, input *CloneInstallsRepoInput) (*plantypes.GitSource, error) {
	var vcsConn app.VCSConnection
	if err := a.db.WithContext(ctx).
		Where(app.VCSConnection{ID: input.VCSConnectionID}).
		First(&vcsConn).Error; err != nil {
		return nil, fmt.Errorf("VCS connection not found: %w", err)
	}

	repoOwner, repoName := parseRepo(input.Repo, vcsConn.GithubAccountName)

	token, err := a.vcsHelpers.CreateInstallationToken(ctx, &vcsConn, repoName)
	if err != nil {
		return nil, fmt.Errorf("unable to create installation token: %w", err)
	}

	ref := input.Branch
	if input.CommitSHA != "" {
		ref = input.CommitSHA
	}

	return &plantypes.GitSource{
		URL: githubpkg.RepoPath(repoOwner, repoName, token),
		Ref: ref,
	}, nil
}

func parseRepo(repo, fallbackOwner string) (owner, name string) {
	if strings.Contains(repo, "/") {
		parts := strings.SplitN(repo, "/", 2)
		return parts[0], parts[1]
	}
	return fallbackOwner, repo
}

type CloneInstallsRepoOutput struct {
	SourceDir string `json:"source_dir"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @local
func (a *Activities) CloneInstallsRepo(ctx context.Context, input *CloneInstallsRepoInput) (*CloneInstallsRepoOutput, error) {
	// TEMP: copy from local mono repo for faster iteration
	srcDir := "/Users/jonmorehouse/nuon/mono"
	return &CloneInstallsRepoOutput{
		SourceDir: srcDir,
	}, nil
}
