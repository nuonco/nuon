package service

import (
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
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

func (b *basicVCSConfigRequest) connectedGithubVCSConfig(parentCmp *app.Component) (*app.ConnectedGithubVCSConfig, error) {
	if b.ConnectedGithubVCSConfig == nil {
		return nil, nil
	}

	pieces := strings.SplitN(b.ConnectedGithubVCSConfig.Repo, "/", 2)
	if len(pieces) != 2 {
		return nil, fmt.Errorf("invalid repo, must be of the format <user-name>/<repo-name>")
	}

	if len(parentCmp.App.Org.VCSConnections) < 1 {
		return nil, fmt.Errorf("org must have at least one vcs connection")
	}

	return &app.ConnectedGithubVCSConfig{
		Repo:            b.ConnectedGithubVCSConfig.Repo,
		RepoName:        pieces[1],
		RepoOwner:       pieces[0],
		Directory:       b.ConnectedGithubVCSConfig.Directory,
		Branch:          b.ConnectedGithubVCSConfig.Branch,
		VCSConnectionID: parentCmp.App.Org.VCSConnections[0].ID,
	}, nil
}

func (b *basicVCSConfigRequest) publicGitVCSConfig() *app.PublicGitVCSConfig {
	if b.PublicGitVCSConfig == nil {
		return nil
	}

	return &app.PublicGitVCSConfig{
		Repo:      b.PublicGitVCSConfig.Repo,
		Directory: b.PublicGitVCSConfig.Directory,
		Branch:    b.PublicGitVCSConfig.Branch,
	}
}
