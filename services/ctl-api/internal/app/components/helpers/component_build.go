package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *Helpers) CreateComponentBuild(ctx context.Context, cmpID string, useLatest bool, gitRef *string) (*app.ComponentBuild, error) {
	var vcsCommit *app.VCSConnectionCommit
	if useLatest {
		var err error
		vcsCommit, err = s.GetComponentCommit(ctx, cmpID)
		if err != nil {
			return nil, fmt.Errorf("unable to get latest commit for connection: %w", err)
		}
	}
	if vcsCommit != nil {
		gitRef = generics.ToPtr(vcsCommit.SHA)
	}

	bld := app.ComponentBuild{
		Status:            "queued",
		StatusDescription: "queued and waiting for runner to pick up",
		GitRef:            gitRef,
	}
	if vcsCommit != nil {
		bld.VCSConnectionCommitID = generics.ToPtr(vcsCommit.ID)
	}

	cmp := app.ComponentConfigConnection{}
	err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(1).
		First(&cmp, "component_id = ?", cmpID).Association("ComponentBuilds").Append(&bld)
	if err != nil {
		return nil, fmt.Errorf("unable to create build for component: %w", err)
	}
	return &bld, nil
}
