package service

import (
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

func (b *basicVCSConfigRequest) connectedGithubVCSConfig(parentCmp *app.Component) *app.ConnectedGithubVCSConfig {
	if b.ConnectedGithubVCSConfig == nil {
		return nil
	}

	return &app.ConnectedGithubVCSConfig{
		Repo:            b.ConnectedGithubVCSConfig.Repo,
		Directory:       b.ConnectedGithubVCSConfig.Directory,
		Branch:          b.ConnectedGithubVCSConfig.Branch,
		GitRef:          b.ConnectedGithubVCSConfig.GitRef,
		VCSConnectionID: parentCmp.App.Org.VCSConnections[0].ID,
	}
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
