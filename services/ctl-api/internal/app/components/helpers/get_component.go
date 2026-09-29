package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (s *Helpers) GetComponent(ctx context.Context, cmpID string) (*app.Component, error) {
	return s.getComponent(ctx, s.db, cmpID)
}

func (s *Helpers) GetComponentInTx(ctx context.Context, tx *gorm.DB, cmpID string) (*app.Component, error) {
	return s.getComponent(ctx, tx, cmpID)
}

func (s *Helpers) getComponent(ctx context.Context, db *gorm.DB, cmpID string) (*app.Component, error) {
	cmp := app.Component{}
	res := db.WithContext(ctx).
		Preload("Org").
		Preload("Org.RunnerGroup").
		Preload("Org.RunnerGroup.Runners").
		Preload("Dependencies").
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithOverrideTable(app.LatestComponentConfigConnectionsViewName))
		}).
		Preload("ComponentConfigs.TerraformModuleComponentConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.TerraformModuleComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigs.HelmComponentConfig").
		Preload("ComponentConfigs.HelmComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.HelmComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigs.DockerBuildComponentConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.DockerBuildComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("ComponentConfigs.ExternalImageComponentConfig").
		Preload("ComponentConfigs.ExternalImageComponentConfig.AWSECRImageConfig").
		Preload("ComponentConfigs.ExternalImageComponentConfig.GCPGARImageConfig").
		Preload("ComponentConfigs.ExternalImageComponentConfig.AzureACRImageConfig").
		Preload("ComponentConfigs.JobComponentConfig").
		Preload("ComponentConfigs.KubernetesManifestComponentConfig").
		Preload("ComponentConfigs.PulumiComponentConfig").
		Preload("ComponentConfigs.PulumiComponentConfig.PublicGitVCSConfig").
		Preload("ComponentConfigs.PulumiComponentConfig.ConnectedGithubVCSConfig").
		Preload("ComponentConfigs.PulumiComponentConfig.ConnectedGithubVCSConfig.VCSConnection").
		First(&cmp, "id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	return &cmp, nil
}
