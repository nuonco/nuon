package activities

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateWarningsRequest struct {
	RunnerID   string `validate:"required"`
	Warnings   pq.StringArray
	IsAliasTag bool
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateWarnings(ctx context.Context, req UpdateWarningsRequest) error {
	runner := app.Runner{
		ID: req.RunnerID,
	}

	updates := map[string]any{
		"warnings": req.Warnings,
	}

	res := a.db.WithContext(ctx).Model(&runner).Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("unable to update runner warnings: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no runner found: %s %w", req.RunnerID, gorm.ErrRecordNotFound)
	}

	// Merge is_alias_tag into the runner's status_v2 metadata.
	if err := generics.MergeJSONBMetadata(a.db.WithContext(ctx), &app.Runner{}, req.RunnerID, "status_v2", map[string]any{
		"is_alias_tag": req.IsAliasTag,
	}); err != nil {
		return fmt.Errorf("unable to update runner metadata: %w", err)
	}

	return nil
}
