package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallComponentIDsRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetInstallComponentIDs(ctx context.Context, req GetInstallComponentIDsRequest) ([]string, error) {
	comps, err := a.helpers.GetInstallComponents(ctx, req.InstallID)
	if err != nil {
		return nil, generics.TemporalGormError(err, "get install components")
	}
	ids := make([]string, 0, len(comps))
	for _, comp := range comps {
		ids = append(ids, comp.ID)
	}
	return ids, nil
}
