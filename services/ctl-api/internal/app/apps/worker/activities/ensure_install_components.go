package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/ensure"
)

type EnsureInstallComponentsRequest struct {
	AppID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field AppID
func (a *Activities) EnsureInstallComponents(ctx context.Context, req *EnsureInstallComponentsRequest) error {
	return ensure.Components(ctx, a.db, req.AppID, nil)
}
