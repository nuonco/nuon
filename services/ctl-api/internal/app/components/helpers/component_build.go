package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *Helpers) CreateComponentBuild(ctx context.Context, cmpID string, useLatest bool, gitRef *string) (*app.ComponentBuild, error) {
	cmp, err := s.GetComponent(ctx, cmpID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	var vcsCommit *app.VCSConnectionCommit
	switch cmp.LatestConfig.VCSConnectionType {
	case app.VCSConnectionTypeConnectedRepo:
		if useLatest {
			var err error
			vcsCommit, err = s.GetComponentCommit(ctx, cmpID)
			if err != nil {
				return nil, fmt.Errorf("unable to get latest commit for connection: %w", err)
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
		return nil, fmt.Errorf("unable to create build for component: %w", err)
	}
	return &bld, nil
}
