package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func (s *Helpers) CreateComponentBuild(ctx context.Context, cmpID string, useLatest bool, gitRef *string) (*app.ComponentBuild, error) {
	return s.createComponentBuild(ctx, "", cmpID, useLatest, gitRef)
}

func (s *Helpers) CreateComponentBuildWithID(ctx context.Context, buildID, cmpID string, useLatest bool, gitRef *string) (*app.ComponentBuild, error) {
	return s.createComponentBuild(ctx, buildID, cmpID, useLatest, gitRef)
}

func DockerBuildUnsupported() stderr.ErrUser {
	return stderr.ErrUser{
		Err: fmt.Errorf("docker_build components have been deprecated"),
		Description: "docker_build components have been deprecated and are no longer supported. " +
			"Use a container_image component to reference a pre-built image instead.",
		Code: "docker_build_deprecated",
	}
}

func (s *Helpers) createComponentBuild(ctx context.Context, buildID, cmpID string, _ bool, gitRef *string) (*app.ComponentBuild, error) {
	cmp, err := s.GetComponent(ctx, cmpID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	if cmp.LatestConfig == nil {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("no config found on component"),
			Description: "please create a component config before building",
		}
	}
	if cmp.Type == app.ComponentTypeDockerBuild {
		return nil, DockerBuildUnsupported()
	}

	switch cmp.LatestConfig.VCSConnectionType {
	case app.VCSConnectionTypePublicRepo:
		gitRef = &cmp.LatestConfig.PublicGitVCSConfig.Branch
	}

	bld := app.ComponentBuild{
		ID:                          buildID,
		Status:                      "queued",
		StatusDescription:           "queued and waiting for runner to pick up",
		GitRef:                      gitRef,
		ComponentConfigConnectionID: cmp.LatestConfig.ID,
	}
	res := s.db.WithContext(ctx).
		Create(&bld)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create build for component: %v", res.Error)
	}

	if err := s.db.WithContext(ctx).
		Model(&app.ComponentConfigConnection{}).
		Where("id = ?", cmp.LatestConfig.ID).
		Update("latest_build_id", bld.ID).Error; err != nil {
		return nil, fmt.Errorf("unable to set latest_build_id on config connection: %w", err)
	}

	return &bld, nil
}
