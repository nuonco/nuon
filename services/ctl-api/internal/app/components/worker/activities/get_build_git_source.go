package activities

import (
	"context"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildsource"
)

type GetBuildGitSource struct {
	BuildID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field BuildID
// @start-to-close-timeout 5m
func (a *Activities) GetBuildGitSource(ctx context.Context, req GetBuildGitSource) (*plantypes.GitSource, error) {
	build, err := a.getComponentBuildWithConfig(ctx, req.BuildID)
	if err != nil {
		return nil, err
	}
	identity, err := buildsource.Component(ctx, a.db, a.vcsHelpers, build)
	if err != nil {
		return nil, err
	}
	identity, err = buildsource.SaveComponent(ctx, a.db, build.ID, identity)
	if err != nil {
		return nil, err
	}
	return buildsource.GitSource(ctx, a.vcsHelpers, build.ComponentConfigConnection.ConnectedGithubVCSConfig, build.ComponentConfigConnection.PublicGitVCSConfig, identity)
}
