package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateAppBranchRunNoConfigChangesInput struct {
	RunID string `json:"run_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) UpdateAppBranchRunNoConfigChanges(ctx context.Context, input *UpdateAppBranchRunNoConfigChangesInput) error {
	res := a.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where(app.AppBranchRun{ID: input.RunID}).
		Updates(app.AppBranchRun{
			NoConfigChanges: true,
			Status:          "success",
		})
	if res.Error != nil {
		return fmt.Errorf("unable to update run: %w", res.Error)
	}

	return nil
}
