package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetReleaseAppRequest struct {
	ReleaseID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ReleaseID
func (a *Activities) GetReleaseApp(ctx context.Context, req *GetReleaseAppRequest) (*app.App, error) {
	app := &app.App{}
	res := a.db.WithContext(ctx).
		Joins("JOIN installs_view_v4 ON installs_view_v4.app_id = apps.id").
		Joins("JOIN component_release_steps ON installs_view_v4.id = ANY (component_release_steps.requested_install_ids)").
		Where("component_release_steps.release_id = ?", req.ReleaseID).
		First(app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return app, nil
}
