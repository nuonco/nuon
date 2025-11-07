package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

type PublicGitVCSConfigRequest struct {
	Repo      string `validate:"required"`
	Directory string `validate:"required"`
	Branch    string `validate:"required"`
}

type ConnectedGithubVCSConfigRequest struct {
	Repo      string `validate:"required"`
	Directory string `validate:"required"`

	Branch string `validate:"required_without=GitRef"`
	GitRef string `validate:"required_without=Branch"`
}

type basicVCSConfigRequest struct {
	PublicGitVCSConfig       *PublicGitVCSConfigRequest       `json:"public_git_vcs_config" validate:"required_if=PublicGitVCSConfig nil"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfigRequest `json:"connected_github_vcs_config" `
}

func (b basicVCSConfigRequest) Validate() error {
	if b.PublicGitVCSConfig != nil && b.ConnectedGithubVCSConfig != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("both public and connected github config set"),
			Description: "only one of connected github or public git configs can be set",
		}
	}

	if b.PublicGitVCSConfig == nil && b.ConnectedGithubVCSConfig == nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("one of public and connected github config set"),
			Description: "one of connected github or public git configs must be set",
			Code:        "vcs_config_required",
		}
	}

	return nil
}

func (b *basicVCSConfigRequest) connectedGithubVCSConfig(ctx context.Context,
	parentCmp *app.Component, vcsHelpers *vcshelpers.Helpers,
) (*app.ConnectedGithubVCSConfig, error) {
	if b.ConnectedGithubVCSConfig == nil {
		return nil, nil
	}

	owner, repo, err := vcsHelpers.SplitRepoSlug(b.ConnectedGithubVCSConfig.Repo)
	if err != nil {
		return nil, err
	}

	vcsConnID, err := vcsHelpers.LookupVCSConnection(ctx, owner, repo, parentCmp.App.Org.VCSConnections)
	if err != nil {
		return nil, err
	}

	return &app.ConnectedGithubVCSConfig{
		Repo:            b.ConnectedGithubVCSConfig.Repo,
		RepoName:        repo,
		RepoOwner:       owner,
		Directory:       b.ConnectedGithubVCSConfig.Directory,
		Branch:          b.ConnectedGithubVCSConfig.Branch,
		VCSConnectionID: vcsConnID,
	}, nil
}

func (b *basicVCSConfigRequest) publicGitVCSConfig(
	ctx context.Context,
	parentCmp *app.Component,
	vcsHelpers *vcshelpers.Helpers,
) (*app.PublicGitVCSConfig, error) {
	if b.PublicGitVCSConfig == nil {
		return nil, nil
	}

	repo := b.PublicGitVCSConfig.Repo
	if strings.HasPrefix("git@", repo) {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("invalid git clone url"),
			Description: "Please use either a <owner>/<repo> format, or a full https:// public clone url",
		}
	}

	owner, repoName, err := vcsHelpers.SplitRepoSlug(repo)
	if err != nil {
		return nil, err
	}

	_, err = vcsHelpers.LookupVCSConnection(ctx, owner, repoName, parentCmp.App.Org.VCSConnections)
	if err == nil {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("repo is connected using vcs connection"),
			Description: "repo is connected using vcs connection, please use connected_repo instead",
		}
	}

	return &app.PublicGitVCSConfig{
		Repo:      repo,
		Directory: b.PublicGitVCSConfig.Directory,
		Branch:    b.PublicGitVCSConfig.Branch,
	}, nil
}
