package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type ForceSandboxModeRequest struct {
	OrgID string `validate:"required"`
}

// @temporal-gen activity
// @by-id OrgID
func (a *Activities) ForceSandboxMode(ctx context.Context, req ForceSandboxModeRequest) error {
	if err := a.forceSandboxMode(ctx, req.OrgID); err != nil {
		return errors.Wrap(err, "unable to force sandbox mode")
	}

	return nil
}

func (a *Activities) forceSandboxMode(ctx context.Context, orgID string) error {
	org := &app.Org{
		ID: orgID,
	}
	res := a.db.WithContext(ctx).
		Model(org).
		Updates(app.Org{
			SandboxMode: true,
			OrgType:     app.OrgTypeSandbox,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to force sandbox mode")
	}

	return nil
}
