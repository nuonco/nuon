package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

func (s *Helpers) GetComponent(ctx context.Context, cmpID string) (*app.Component, error) {
	cmp := app.Component{}
	res := s.db.WithContext(ctx).
		Preload("ComponentConfigs").
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("component_config_connections.created_at DESC")
		}).

		// preload all terraform configs
		Preload("ComponentConfigs.TerraformModuleComponentConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all helm configs
		Preload("ComponentConfigs.HelmComponentConfig").
		Preload("ComponentConfigs.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.HelmComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all docker configs
		Preload("ComponentConfigs.DockerBuildComponentConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.ConnectedGithubVCSConfig.VCSConnection").

		// preload all external image configs
		Preload("ComponentConfigs.ExternalImageComponentConfig").

		// preload all job configs
		Preload("ComponentConfigs.JobComponentConfig").
		First(&cmp, "id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	return &cmp, nil
}
