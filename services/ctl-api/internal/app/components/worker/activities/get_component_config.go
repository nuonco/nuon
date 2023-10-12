package activities

import (
	"context"
	"fmt"

	componentsv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRequest struct {
	BuildID string `validate:"required"`
}

func (a *Activities) GetComponentConfig(ctx context.Context, req GetRequest) (*componentsv1.Component, error) {
	bld := app.ComponentBuild{}
	res := a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("ComponentConfigConnection").
		Preload("ComponentConfigConnection.Component").

		// preload all terraform configs
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig").
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all helm configs
		Preload("ComponentConfigConnection.HelmComponentConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all docker configs
		Preload("ComponentConfigConnection.DockerBuildComponentConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.BasicDeployConfig").

		// preload all external image configs
		Preload("ComponentConfigConnection.ExternalImageComponentConfig").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.BasicDeployConfig").

		// preload all job configs
		Preload("ComponentConfigConnection.JobComponentConfig").

		// get config by build ID
		First(&bld, "id = ?", req.BuildID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get build: %w", res.Error)
	}

	compCfg, err := a.components.FromBuild(&bld)
	if err != nil {
		return nil, fmt.Errorf("unable to convert build to component config: %w", err)
	}

	return compCfg, nil
}
