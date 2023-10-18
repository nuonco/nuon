package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v50/github"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"golang.org/x/oauth2"
)

func (s *service) getComponentConnectionCommit(ctx context.Context, cmpID string) (*app.VCSConnectionCommit, error) {
	cmp := app.ComponentConfigConnection{}
	res := s.db.WithContext(ctx).
		Preload("Component", "id = ?", cmpID).
		Order("created_at DESC").

		// preload all terraform configs
		Preload("TerraformModuleComponentConfig").
		Preload("TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("TerraformModuleComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all helm configs
		Preload("HelmComponentConfig").
		Preload("HelmComponentConfig.PublicGitVCSConfig").
		Preload("HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("HelmComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all docker configs
		Preload("DockerBuildComponentConfig").
		Preload("DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("DockerBuildComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("DockerBuildComponentConfig.BasicDeployConfig").

		// preload all external image configs
		Preload("ExternalImageComponentConfig").
		Preload("ExternalImageComponentConfig.PublicGitVCSConfig").
		Preload("ExternalImageComponentConfig.ConnectedGithubVCSConfig").
		Preload("ExternalImageComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ExternalImageComponentConfig.BasicDeployConfig").

		// preload
		First(&cmp, "component_id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component config connection: %w", res.Error)
	}

	// check for the correct component configuration
	var connectedGithubVCSConfig *app.ConnectedGithubVCSConfig
	if cmp.TerraformModuleComponentConfig != nil {
		connectedGithubVCSConfig = cmp.TerraformModuleComponentConfig.ConnectedGithubVCSConfig
	} else if cmp.HelmComponentConfig != nil {
		connectedGithubVCSConfig = cmp.HelmComponentConfig.ConnectedGithubVCSConfig
	} else if cmp.DockerBuildComponentConfig != nil {
		connectedGithubVCSConfig = cmp.DockerBuildComponentConfig.ConnectedGithubVCSConfig
	}
	if connectedGithubVCSConfig == nil {
		return nil, nil
	}

	// find the latest commit for this connection
	commit, err := s.getVCSConfigLatestCommit(ctx, connectedGithubVCSConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to get the latest commit: %w", err)
	}
	vcsCommit := app.VCSConnectionCommit{
		SHA:             *commit.SHA,
		Message:         *commit.Commit.Message,
		VCSConnectionID: connectedGithubVCSConfig.VCSConnectionID,
		AuthorName:      generics.FromPtrStr(commit.Author.Name),
		AuthorEmail:     generics.FromPtrStr(commit.Author.Email),
	}
	res = s.db.WithContext(ctx).Create(&vcsCommit)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create vcs commit: %w", res.Error)
	}

	return &vcsCommit, nil
}

func (s *service) getVCSConfigLatestCommit(ctx context.Context, vcsCfg *app.ConnectedGithubVCSConfig) (*github.RepositoryCommit, error) {
	// get a static token
	installID, err := strconv.ParseInt(vcsCfg.VCSConnection.GithubInstallID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to get install ID: %w", err)
	}
	resp, _, err := s.ghClient.Apps.CreateInstallationToken(ctx, installID, &github.InstallationTokenOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get installation token: %w", err)
	}

	// get a client with the github install token
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *resp.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	commitResp, _, err := client.Repositories.GetCommit(ctx, vcsCfg.RepoOwner, vcsCfg.RepoName, vcsCfg.Branch, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get latest commit: %w", err)
	}

	return commitResp, nil
}
