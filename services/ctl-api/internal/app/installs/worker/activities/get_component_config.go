package activities

import (
	"context"
	"fmt"

	componentsv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentConfigRequest struct {
	DeployID string `validate:"required"`
}

func (a *Activities) GetComponentConfig(ctx context.Context, req GetComponentConfigRequest) (*componentsv1.Component, error) {
	// NOTE(jm): we have to load a TON of stuff here, but most of this complexity comes from the fact that the VCS
	// config is a child of the component config, not the component version.
	dep := app.InstallDeploy{}
	res := a.db.WithContext(ctx).
		Preload("InstallComponent.Install").
		Preload("InstallComponent.Install.App").
		Preload("InstallComponent.Install.App.Org").

		// build
		Preload("ComponentBuild.VCSConnectionCommit").
		Preload("ComponentBuild.ComponentConfigConnection").
		Preload("ComponentBuild.ComponentConfigConnection.Component").

		// preload all terraform configs
		Preload("ComponentBuild.ComponentConfigConnection.TerraformModuleComponentConfig").
		Preload("ComponentBuild.ComponentConfigConnection.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentBuild.ComponentConfigConnection.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").

		// preload all helm configs
		Preload("ComponentBuild.ComponentConfigConnection.HelmComponentConfig").
		Preload("ComponentBuild.ComponentConfigConnection.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentBuild.ComponentConfigConnection.HelmComponentConfig.ConnectedGithubVCSConfig").

		// preload all docker configs
		Preload("ComponentBuild.ComponentConfigConnection.DockerBuildComponentConfig").
		Preload("ComponentBuild.ComponentConfigConnection.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentBuild.ComponentConfigConnection.DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentBuild.ComponentConfigConnection.DockerBuildComponentConfig.BasicDeployConfig").

		// preload all external image configs
		Preload("ComponentBuild.ComponentConfigConnection.ExternalImageComponentConfig").
		Preload("ComponentBuild.ComponentConfigConnection.ExternalImageComponentConfig.PublicGitVCSConfig").
		Preload("ComponentBuild.ComponentConfigConnection.ExternalImageComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentBuild.ComponentConfigConnection.ExternalImageComponentConfig.BasicDeployConfig").
		First(&dep, "id = ?", req.DeployID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get deploy: %w", res.Error)
	}

	// get sibling deploys InstallDeploy
	var installDeploys []app.InstallDeploy
	res = a.db.WithContext(ctx).
		Find(&dep, "id = ?", req.DeployID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploys: %w", res.Error)
	}

	compCfg, err := a.components.FromDeploy(&dep, installDeploys)
	if err != nil {
		return nil, fmt.Errorf("unable to convert deploy to component config: %w", err)
	}

	return compCfg, nil
}
