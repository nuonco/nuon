package activities

import (
	"context"

	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetBuildGitSource struct {
	BuildID string `validate:"required"`
}

// @temporal-gen activity
// @by-id BuildID
func (a *Activities) GetBuildGitSource(ctx context.Context, req GetBuildGitSource) (*plantypes.GitSource, error) {
	build, err := a.getComponentBuildWithConfig(ctx, req.BuildID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get build config")
	}

	switch build.ComponentConfigConnection.VCSConnectionType {
	case app.VCSConnectionTypeConnectedRepo:
		return a.vcsHelpers.GetGitSource(ctx, build.ComponentConfigConnection.ConnectedGithubVCSConfig)
	case app.VCSConnectionTypePublicRepo:
		return a.vcsHelpers.GetPubliGitSource(ctx, build.ComponentConfigConnection.PublicGitVCSConfig)
	default:
	}

	return nil, nil
}
