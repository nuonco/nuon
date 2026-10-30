package helpers

import (
	"context"
	"encoding/json"

	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func (h *Helpers) MarkInstallStateStale(ctx context.Context, installID string) error {
	return h.MarkInstallStatePartialsStale(ctx, installID, pkgstate.AllPartials...)
}

func (h *Helpers) MarkInstallStatePartialsStale(ctx context.Context, installID string, partials ...pkgstate.PartialName) error {
	if len(partials) == 0 {
		return nil
	}
	data, err := json.Marshal(partials)
	if err != nil {
		return err
	}
	return h.db.WithContext(ctx).Exec("SELECT invalidate_install_state(ARRAY[?]::text[], ?::jsonb)", installID, string(data)).Error
}
