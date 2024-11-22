package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

type PublicGitVCSSandboxConfigRequest struct {
	Repo      string `validate:"required"`
	Directory string `validate:"required"`
	Branch    string `validate:"required"`
}

type ConnectedGithubVCSSandboxConfigRequest struct {
	Repo      string `validate:"required"`
	Directory string `validate:"required"`

	Branch string `validate:"required_without=GitRef"`
	GitRef string `validate:"required_without=Branch"`
}

type basicVCSConfigRequest struct {
	PublicGitVCSConfig       *PublicGitVCSSandboxConfigRequest       `json:"public_git_vcs_config" validate:"required_if=PublicGitVCSConfig nil"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSSandboxConfigRequest `json:"connected_github_vcs_config" `
}

func (b *basicVCSConfigRequest) connectedGithubVCSConfig(ctx context.Context, parentApp *app.App, vcsHelpers *vcshelpers.Helpers) (*app.ConnectedGithubVCSConfig, error) {
	if b.ConnectedGithubVCSConfig == nil {
		return nil, nil
	}

	owner, repo, err := vcsHelpers.SplitRepoSlug(b.ConnectedGithubVCSConfig.Repo)
	if err != nil {
		return nil, err
	}

	vcsConnID, err := vcsHelpers.LookupVCSConnection(ctx, owner, repo, parentApp.Org.VCSConnections)
	if err != nil {
		return nil, err
	}

	return &app.ConnectedGithubVCSConfig{
		Repo:            b.ConnectedGithubVCSConfig.Repo,
		RepoName:        owner,
		RepoOwner:       repo,
		Directory:       b.ConnectedGithubVCSConfig.Directory,
		Branch:          b.ConnectedGithubVCSConfig.Branch,
		VCSConnectionID: vcsConnID,
	}, nil
}

func (b *basicVCSConfigRequest) publicGitVCSConfig() (*app.PublicGitVCSConfig, error) {
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

	return &app.PublicGitVCSConfig{
		Repo:      repo,
		Directory: b.PublicGitVCSConfig.Directory,
		Branch:    b.PublicGitVCSConfig.Branch,
	}, nil
}
