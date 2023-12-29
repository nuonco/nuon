package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// GetComponentCommit will return a commit for a component, when a connected git source is attached.
func (s *Helpers) GetComponentCommit(ctx context.Context, cmpID string) (*app.VCSConnectionCommit, error) {
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

		// preload all external image configs
		Preload("ExternalImageComponentConfig").

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
	commit, err := s.vcsHelpers.GetVCSConfigLatestCommit(ctx, connectedGithubVCSConfig)
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
