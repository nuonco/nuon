package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetHelmComponentConfigRequest struct {
	ComponentConfigConnectionID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ComponentConfigConnectionID
func (a *Activities) GetHelmComponentConfig(ctx context.Context, req GetHelmComponentConfigRequest) (*app.HelmComponentConfig, error) {
	return a.getHelmComponentConfig(ctx, req.ComponentConfigConnectionID)
}

func (a *Activities) getHelmComponentConfig(ctx context.Context, connectionID string) (*app.HelmComponentConfig, error) {
	var config app.HelmComponentConfig

	res := a.db.WithContext(ctx).
		Preload("PublicGitVCSConfig").
		Preload("ConnectedGithubVCSConfig").
		Preload("ConnectedGithubVCSConfig.VCSConnection").
		First(&config, "component_config_connection_id = ?", connectionID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get helm component config")
	}

	return &config, nil
}
