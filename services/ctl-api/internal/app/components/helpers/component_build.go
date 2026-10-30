package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildselection"
)

func (s *Helpers) CreateComponentBuild(ctx context.Context, cmpID string, useLatest bool, gitRef *string) (*app.ComponentBuild, error) {
	return s.createComponentBuild(ctx, "", cmpID, "", useLatest, gitRef, nil)
}

func (s *Helpers) CreateComponentBuildWithRequest(ctx context.Context, cmpID string, useLatest bool, gitRef *string, wfRequest *app.WorkflowRequest) (*app.ComponentBuild, error) {
	return s.createComponentBuild(ctx, "", cmpID, "", useLatest, gitRef, wfRequest)
}

func (s *Helpers) CreateComponentBuildWithID(ctx context.Context, buildID, cmpID string, useLatest bool, gitRef *string) (*app.ComponentBuild, error) {
	return s.createComponentBuild(ctx, buildID, cmpID, "", useLatest, gitRef, nil)
}

// CreateComponentBuildForConfigConnection creates a build attached to a specific
// CCC (e.g. a branch run's app-config CCC) instead of the global LatestConfig.
func (s *Helpers) CreateComponentBuildForConfigConnection(ctx context.Context, cmpID, componentConfigConnectionID string, gitRef *string) (*app.ComponentBuild, error) {
	return s.createComponentBuild(ctx, "", cmpID, componentConfigConnectionID, false, gitRef, nil)
}

func DockerBuildUnsupported() stderr.ErrUser {
	return stderr.ErrUser{
		Err: fmt.Errorf("docker_build components have been deprecated"),
		Description: "docker_build components have been deprecated and are no longer supported. " +
			"Use a container_image component to reference a pre-built image instead.",
		Code: "docker_build_deprecated",
	}
}

func (s *Helpers) createComponentBuild(ctx context.Context, buildID, cmpID, componentConfigConnectionID string, _ bool, gitRef *string, wfRequest *app.WorkflowRequest) (*app.ComponentBuild, error) {
	cmp, err := s.GetComponent(ctx, cmpID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	if cmp.Type == app.ComponentTypeDockerBuild {
		return nil, DockerBuildUnsupported()
	}

	configConn, err := s.resolveBuildConfigConnection(ctx, cmp, componentConfigConnectionID)
	if err != nil {
		return nil, err
	}

	if gitRef == nil && configConn.VCSConnectionType == app.VCSConnectionTypePublicRepo && configConn.PublicGitVCSConfig != nil {
		gitRef = &configConn.PublicGitVCSConfig.Branch
	}

	bld := app.ComponentBuild{
		ID:                          buildID,
		Status:                      app.ComponentBuildStatusQueued,
		StatusDescription:           "queued and waiting for runner to pick up",
		GitRef:                      gitRef,
		ComponentConfigConnectionID: configConn.ID,
		Checksum:                    configConn.Checksum,
		StoredComponentID:           cmp.ID,
		Request:                     wfRequest,
	}
	create := func(tx *gorm.DB, cfg *app.AppConfig) error {
		if cfg.BuildsPinned && componentConfigConnectionID == "" {
			return stderr.ErrUser{
				Err:         fmt.Errorf("manual builds cannot change branch-run app configs"),
				Description: "Trigger an app branch run to rebuild this component.",
			}
		}
		if err := tx.Create(&bld).Error; err != nil {
			return fmt.Errorf("unable to create build for component: %w", err)
		}
		if err := tx.Model(&app.ComponentConfigConnection{}).
			Where(app.ComponentConfigConnection{ID: configConn.ID}).
			Update("latest_build_id", bld.ID).Error; err != nil {
			return fmt.Errorf("unable to set latest_build_id on config connection: %w", err)
		}
		return nil
	}
	if configConn.AppConfigID == "" {
		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return create(tx, &app.AppConfig{}) })
	} else {
		err = buildselection.WithMutableConfig(ctx, s.db, configConn.AppConfigID, create)
	}
	if err != nil {
		return nil, err
	}

	return &bld, nil
}

func (s *Helpers) resolveBuildConfigConnection(ctx context.Context, cmp *app.Component, componentConfigConnectionID string) (*app.ComponentConfigConnection, error) {
	query := s.db.WithContext(ctx).
		Preload("TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("HelmComponentConfig.PublicGitVCSConfig").
		Preload("DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("KubernetesManifestComponentConfig.PublicGitVCSConfig").
		Preload("PulumiComponentConfig.PublicGitVCSConfig").
		Where(app.ComponentConfigConnection{ID: componentConfigConnectionID, ComponentID: cmp.ID})
	if componentConfigConnectionID == "" {
		if cmp.LatestConfig == nil {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("no config found on component"),
				Description: "please create a component config before building",
			}
		}
		if cmp.LatestConfig.AppConfigID == "" {
			return cmp.LatestConfig, nil
		}
		legacyConfigs := s.db.WithContext(ctx).Model(&app.AppConfig{}).Select("id").Where(map[string]any{"builds_pinned": false})
		query = query.
			Where("(app_config_id IN (?) OR app_config_id IS NULL OR app_config_id = '')", legacyConfigs).
			Order("created_at DESC, id DESC")
	}

	var ccc app.ComponentConfigConnection
	err := query.First(&ccc).Error
	if err != nil {
		if componentConfigConnectionID == "" && errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("manual builds cannot change branch-run app configs"),
				Description: "Trigger an app branch run to rebuild this component.",
			}
		}
		return nil, fmt.Errorf("unable to get component config connection: %w", err)
	}
	return &ccc, nil
}
