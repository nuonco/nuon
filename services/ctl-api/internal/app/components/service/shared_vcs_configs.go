package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
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

func (b *basicVCSConfigRequest) lookupVCSConnection(ctx context.Context,
	ghClient *github.Client,
	owner, name string,
	vcsConnections []app.VCSConnection) (string, error) {
	if len(vcsConnections) < 1 {
		return "", stderr.ErrUser{
			Err:         fmt.Errorf("no vcs connections on org: %w", gorm.ErrRecordNotFound),
			Description: "please create a vcs connection before proceeding",
		}
	}

	for _, vcsConn := range vcsConnections {
		installID, err := strconv.ParseInt(vcsConn.GithubInstallID, 10, 64)
		if err != nil {
			return "", fmt.Errorf("unable to get install ID: %w", err)
		}

		resp, _, err := ghClient.Apps.CreateInstallationToken(ctx, installID, &github.InstallationTokenOptions{})
		if err != nil {
			return "", fmt.Errorf("unable to get installation token: %w", err)
		}

		// get a client with the github install token
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *resp.Token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)

		repo, _, err := client.Repositories.Get(ctx, owner, name)
		if err != nil {
			continue
		}

		if *repo.Visibility == "public" {
			return "", stderr.ErrUser{
				Err:         fmt.Errorf("can not use a public repo with a connected_repo config"),
				Description: "please use a `public_repo` block instead",
			}
		}
		return vcsConn.ID, nil
	}

	return "", stderr.ErrUser{
		Err:         fmt.Errorf("no vcs connection found with access to %s/%s", owner, name),
		Description: "please make sure vcs connection has access to this repo",
	}
}

func (b *basicVCSConfigRequest) connectedGithubVCSConfig(ctx context.Context, parentCmp *app.Component, ghClient *github.Client) (*app.ConnectedGithubVCSConfig, error) {
	if b.ConnectedGithubVCSConfig == nil {
		return nil, nil
	}

	pieces := strings.SplitN(b.ConnectedGithubVCSConfig.Repo, "/", 2)
	if len(pieces) != 2 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("invalid repo, must be of the format <user-name>/<repo-name>"),
			Description: "please correct format and try again",
		}
	}

	vcsConnID, err := b.lookupVCSConnection(ctx, ghClient, pieces[0], pieces[1], parentCmp.App.Org.VCSConnections)
	if err != nil {
		return nil, err
	}

	return &app.ConnectedGithubVCSConfig{
		Repo:            b.ConnectedGithubVCSConfig.Repo,
		RepoName:        pieces[1],
		RepoOwner:       pieces[0],
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
