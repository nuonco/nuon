package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetInstallsRequest struct {
	ID string `validate:"required"`
}

// OrgInstall is a bare projection rather than app.Install so scanning does not run
// Install's AfterQuery hook, which fires a per-row org lookup to populate derived
// fields no fan-out needs.
type OrgInstall struct {
	ID string `json:"id"`
}

// GetInstalls returns the IDs of an org's installs, for signals that fan work out
// across them.
//
// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) GetInstalls(ctx context.Context, req GetInstallsRequest) ([]OrgInstall, error) {
	installs := []OrgInstall{}

	res := a.db.WithContext(ctx).
		Model(&app.Install{}).
		Select("id").
		Where(app.Install{OrgID: req.ID}).
		Order("created_at asc").
		Find(&installs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get installs: %w", res.Error)
	}

	return installs, nil
}
