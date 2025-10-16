package activities

import (
	"context"

	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetSandboxRunGitSourceRequest struct {
	AppConfigID string `validate:"required"`
}

// @temporal-gen activity
// @by-id AppConfigID
func (a *Activities) GetSandboxRunGitSource(ctx context.Context, req GetSandboxRunGitSourceRequest) (*plantypes.GitSource, error) {
	cfg, err := a.appsHelpers.GetFullAppConfig(ctx, req.AppConfigID, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	switch cfg.SandboxConfig.VCSConnectionType {
	case app.VCSConnectionTypeConnectedRepo:
		return a.vcsHelpers.GetGitSource(ctx, cfg.SandboxConfig.ConnectedGithubVCSConfig)
	case app.VCSConnectionTypePublicRepo:
		return a.vcsHelpers.GetPubliGitSource(ctx, cfg.SandboxConfig.PublicGitVCSConfig)
	default:
	}

	return nil, errors.New("no vcs connection found")
}
