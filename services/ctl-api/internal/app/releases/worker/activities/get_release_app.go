package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

type GetReleaseAppRequest struct {
	ReleaseID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ReleaseID
func (a *Activities) GetReleaseApp(ctx context.Context, req *GetReleaseAppRequest) (*app.App, error) {
	respApp := &app.App{}
	installTableOrView := views.TableOrViewName(a.db, &app.Install{}, "")
	res := a.db.WithContext(ctx).
		Joins(fmt.Sprintf("JOIN %s ON %s.app_id = apps.id", installTableOrView, installTableOrView)).
		Joins(fmt.Sprintf("JOIN component_release_steps ON %s.id = ANY (component_release_steps.requested_install_ids)", installTableOrView)).
		Where("component_release_steps.release_id = ?", req.ReleaseID).
		First(respApp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return respApp, nil
}
