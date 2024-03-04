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
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig.ComponentConfigConnection").

		// preload all helm configs
		Preload("ComponentConfigConnection.HelmComponentConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigConnection.HelmComponentConfig.ComponentConfigConnection").

		// preload all docker configs
		Preload("ComponentConfigConnection.DockerBuildComponentConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig.ComponentConfigConnection").

		// preload all external image configs
		Preload("ComponentConfigConnection.ExternalImageComponentConfig").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.ComponentConfigConnection").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.AWSECRImageConfig").

		// preload all job configs
		Preload("ComponentConfigConnection.JobComponentConfig").
		Preload("ComponentConfigConnection.JobComponentConfig.ComponentConfigConnection").

		// get config by build ID
		First(&bld, "id = ?", req.BuildID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get build: %w", res.Error)
	}

	compCfg, err := a.protos.FromBuild(&bld)
	if err != nil {
		return nil, fmt.Errorf("unable to convert build to component config: %w", err)
	}

	if err := compCfg.Validate(); err != nil {
		return nil, fmt.Errorf("component config was invalid: %w", err)
	}

	return compCfg, nil
}
