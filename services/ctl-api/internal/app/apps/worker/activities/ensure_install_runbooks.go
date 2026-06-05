package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/ensure"
)

type EnsureInstallRunbooksRequest struct {
	AppID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field AppID
func (a *Activities) EnsureInstallRunbooks(ctx context.Context, req *EnsureInstallRunbooksRequest) error {
	return ensure.Runbooks(ctx, a.db, req.AppID, nil)
}
