package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
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
	return &bld, nil
}

func (s *Helpers) GetComponentLatestBuild(ctx context.Context, cmpID string) (*app.ComponentBuild, error) {
	cmp := app.Component{}

	// query all builds that belong to the component id, starting at the component to ensure the component exists
	// via the double join.
	res := s.db.WithContext(ctx).
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(s.db, &app.ComponentConfigConnection{}, ".created_at DESC"))
		}).
		Preload("ComponentConfigs.ComponentBuilds", func(db *gorm.DB) *gorm.DB {
			return db.Order("component_builds.created_at DESC").Limit(1)
		}).
		Preload("ComponentConfigs.ComponentBuilds.VCSConnectionCommit").
		First(&cmp, "id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	// pull out the first (and only) component build
	for _, cfg := range cmp.ComponentConfigs {
		for _, bld := range cfg.ComponentBuilds {
			return &bld, nil
		}
	}

	return nil, fmt.Errorf("no build found for component: %w", gorm.ErrRecordNotFound)
}
