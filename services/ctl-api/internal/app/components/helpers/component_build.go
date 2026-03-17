package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func (s *Helpers) CreateComponentBuild(ctx context.Context, cmpID string, useLatest bool, gitRef *string) (*app.ComponentBuild, error) {
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

	var vcsCommit *app.VCSConnectionCommit
	switch cmp.LatestConfig.VCSConnectionType {
	case app.VCSConnectionTypeConnectedRepo:
		if useLatest {
			var err error
			vcsCommit, err = s.GetComponentCommit(ctx, cmpID)
			if err != nil {
				return nil, err
			}

			gitRef = generics.ToPtr(vcsCommit.SHA)
		}
	case app.VCSConnectionTypePublicRepo:
		gitRef = generics.ToPtr(cmp.LatestConfig.PublicGitVCSConfig.Branch)
	}

	bld := app.ComponentBuild{
		Status:                      "queued",
		StatusDescription:           "queued and waiting for runner to pick up",
		GitRef:                      gitRef,
		ComponentConfigConnectionID: cmp.LatestConfig.ID,
	}
	if vcsCommit != nil {
		bld.VCSConnectionCommitID = generics.ToPtr(vcsCommit.ID)
	}

	res := s.db.WithContext(ctx).
		Create(&bld)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create build for component: %v", res.Error)
	}

	// Check for duplicate builds (same commit SHA and config checksum).
	// This is a warning only — the build still proceeds.
	if vcsCommit != nil {
		if dupBuild, err := s.findDuplicateBuild(ctx, bld.ID, cmpID, vcsCommit.SHA, cmp.LatestConfig.Checksum); err == nil && dupBuild != nil {
			bld.StatusV2.Metadata = map[string]any{
				"duplicate_build":       true,
				"duplicate_of_build_id": dupBuild.ID,
			}
			s.db.WithContext(ctx).Model(&bld).Update("status_v2", bld.StatusV2)
		}
	}

	return &bld, nil
}

func (s *Helpers) findDuplicateBuild(ctx context.Context, excludeBuildID string, componentID string, commitSHA string, checksum string) (*app.ComponentBuild, error) {
	var build app.ComponentBuild
	res := s.db.WithContext(ctx).
		Joins("JOIN vcs_connection_commits ON vcs_connection_commits.id = component_builds.vcs_connection_commit_id").
		Joins("JOIN component_config_connections ON component_config_connections.id = component_builds.component_config_connection_id").
		Where("component_builds.id != ?", excludeBuildID).
		Where("component_config_connections.component_id = ?", componentID).
		Where("vcs_connection_commits.sha = ?", commitSHA).
		Where("component_config_connections.checksum = ?", checksum).
		Order("component_builds.created_at DESC").
		First(&build)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return &build, nil
}
