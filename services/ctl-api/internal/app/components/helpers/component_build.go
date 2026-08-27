package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

func (s *Helpers) CreateComponentBuild(ctx context.Context, cmpID string, useLatest bool, gitRef *string) (*app.ComponentBuild, error) {
	return s.createComponentBuild(ctx, s.db, "", cmpID, "", useLatest, gitRef)
}

func (s *Helpers) CreateComponentBuildWithID(ctx context.Context, buildID, cmpID string, useLatest bool, gitRef *string) (*app.ComponentBuild, error) {
	return s.createComponentBuild(ctx, s.db, buildID, cmpID, "", useLatest, gitRef)
}

// CreateComponentBuildForConfigConnection creates a build attached to a specific
// CCC (e.g. a branch run's app-config CCC) instead of the global LatestConfig.
func (s *Helpers) CreateComponentBuildForConfigConnection(ctx context.Context, cmpID, componentConfigConnectionID string, gitRef *string) (*app.ComponentBuild, error) {
	return s.createComponentBuild(ctx, s.db, "", cmpID, componentConfigConnectionID, false, gitRef)
}

// CreateComponentBuildInTx creates the build through the caller's transaction.
// When componentConfigConnectionID is set, the build is attached to that CCC.
func (s *Helpers) CreateComponentBuildInTx(ctx context.Context, tx *gorm.DB, cmpID, componentConfigConnectionID string, useLatest bool, gitRef *string) (*app.ComponentBuild, error) {
	return s.createComponentBuild(ctx, tx, "", cmpID, componentConfigConnectionID, useLatest, gitRef)
}

func DockerBuildUnsupported() stderr.ErrUser {
	return stderr.ErrUser{
		Err: fmt.Errorf("docker_build components have been deprecated"),
		Description: "docker_build components have been deprecated and are no longer supported. " +
			"Use a container_image component to reference a pre-built image instead.",
		Code: "docker_build_deprecated",
	}
}

func (s *Helpers) createComponentBuild(ctx context.Context, db *gorm.DB, buildID, cmpID, componentConfigConnectionID string, _ bool, gitRef *string) (*app.ComponentBuild, error) {
	cmp, err := s.getComponent(ctx, db, cmpID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	if cmp.Type == app.ComponentTypeDockerBuild {
		return nil, DockerBuildUnsupported()
	}

	configConn, err := s.resolveBuildConfigConnection(ctx, db, cmp, componentConfigConnectionID)
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
	}
	res := db.WithContext(ctx).
		Create(&bld)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create build for component: %v", res.Error)
	}

	if err := db.WithContext(ctx).
		Model(&app.ComponentConfigConnection{}).
		Where("id = ?", configConn.ID).
		Update("latest_build_id", bld.ID).Error; err != nil {
		return nil, fmt.Errorf("unable to set latest_build_id on config connection: %w", err)
	}

	return &bld, nil
}

func (s *Helpers) resolveBuildConfigConnection(ctx context.Context, db *gorm.DB, cmp *app.Component, componentConfigConnectionID string) (*app.ComponentConfigConnection, error) {
	if componentConfigConnectionID == "" {
		if cmp.LatestConfig == nil {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("no config found on component"),
				Description: "please create a component config before building",
			}
		}
		return cmp.LatestConfig, nil
	}

	var ccc app.ComponentConfigConnection
	res := db.WithContext(ctx).
		Preload("TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("HelmComponentConfig.PublicGitVCSConfig").
		Preload("DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("KubernetesManifestComponentConfig.PublicGitVCSConfig").
		Preload("PulumiComponentConfig.PublicGitVCSConfig").
		Where(app.ComponentConfigConnection{ID: componentConfigConnectionID, ComponentID: cmp.ID}).
		First(&ccc)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component config connection: %w", res.Error)
	}
	return &ccc, nil
}
