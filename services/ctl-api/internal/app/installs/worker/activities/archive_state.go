package activities

import (
	"context"
)

type ArchiveStateRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) ArchiveState(ctx context.Context, req *ArchiveStateRequest) error {
	return a.helpers.ArchiveInstallState(ctx, req.InstallID)
}
