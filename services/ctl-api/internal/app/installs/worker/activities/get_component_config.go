package activities

import (
	"context"
	"fmt"

	componentsv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
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
		Preload("InstallComponent.Install.InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs.created_at DESC")
		}).
		Preload("InstallComponent.Install.InstallInputs.AppInputConfig").
		Preload("InstallComponent.Install.InstallInputs.AppInputConfig.AppInputs").
		Preload("InstallComponent.Install.App").
		Preload("InstallComponent.Install.App.Org").

		// build
		Preload("ComponentBuild.VCSConnectionCommit").
		Preload("ComponentBuild.ComponentConfigConnection").
		Preload("ComponentBuild.ComponentConfigConnection.Component").

		// preload all terraform configs
		Preload("ComponentBuild.ComponentConfigConnection.TerraformModuleComponentConfig").
		Preload("ComponentBuild.ComponentConfigConnection.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentBuild.ComponentConfigConnection.TerraformModuleComponentConfig.ComponentConfigConnection").
		Preload("ComponentBuild.ComponentConfigConnection.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").

		// preload all helm configs
		Preload("ComponentBuild.ComponentConfigConnection.HelmComponentConfig").
		Preload("ComponentBuild.ComponentConfigConnection.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentBuild.ComponentConfigConnection.HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentBuild.ComponentConfigConnection.HelmComponentConfig.ComponentConfigConnection").

		// preload all docker configs
		Preload("ComponentBuild.ComponentConfigConnection.DockerBuildComponentConfig").
		Preload("ComponentBuild.ComponentConfigConnection.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentBuild.ComponentConfigConnection.DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentBuild.ComponentConfigConnection.DockerBuildComponentConfig.ComponentConfigConnection").

		// preload all external image configs
		Preload("ComponentBuild.ComponentConfigConnection.ExternalImageComponentConfig").
		Preload("ComponentBuild.ComponentConfigConnection.ExternalImageComponentConfig.ComponentConfigConnection").

		// preload all job configs
		Preload("ComponentBuild.ComponentConfigConnection.JobComponentConfig").
		Preload("ComponentBuild.ComponentConfigConnection.JobComponentConfig.ComponentConfigConnection").

		// get config by deploy ID
		First(&dep, "id = ?", req.DeployID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get deploy: %w", res.Error)
	}

	// get the latest deploy for each sibling component, by doing a distinct grouping of all deploys by
	// install_component parent (and grabbing latest).
	//
	// downstream, the component adapters need the component object itself, so preload these.
	var deploys []app.InstallDeploy
	res = a.db.WithContext(ctx).
		Preload("InstallComponent.Component").
		Select("DISTINCT ON (install_component_id) install_component_id, install_deploys.id", "install_deploys.created_at").
		Order("install_component_id").
		Order("install_deploys.created_at desc").
		Joins("JOIN install_components ON install_components.id=install_component_id").
		Where("status = 'active'").
		Not("install_components.component_id = ?", dep.InstallComponent.ComponentID).
		Find(&deploys, "install_id = ?", dep.InstallComponent.InstallID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploys: %w", res.Error)
	}

	compCfg, err := a.components.FromDeploy(&dep, deploys)
	if err != nil {
		return nil, fmt.Errorf("unable to convert deploy to component config: %w", err)
	}

	if err := compCfg.Validate(); err != nil {
		return nil, fmt.Errorf("component config was invalid: %w", err)
	}

	return compCfg, nil
}
