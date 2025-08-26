package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type MarkStateStaleRequest struct {
	TriggeredByID   string
	TriggeredByType string

	InstallID string
}

// @temporal-gen activity
func (a *Activities) MarkStateStale(ctx context.Context, req *MarkStateStaleRequest) error {
	if err := a.helpers.MarkInstallStateStale(ctx, req.InstallID); err != nil {
		return generics.TemporalGormError(err)
	}

	return nil
}
