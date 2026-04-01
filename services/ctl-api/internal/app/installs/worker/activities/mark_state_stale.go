package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type MarkStateStaleRequest struct {
	TriggeredByID   string
	TriggeredByType string
	InstallID       string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) MarkStateStale(ctx context.Context, req *MarkStateStaleRequest) error {
	if err := a.helpers.MarkInstallStateStale(ctx, req.InstallID); err != nil {
		return generics.TemporalGormError(err)
	}

	return nil
}
