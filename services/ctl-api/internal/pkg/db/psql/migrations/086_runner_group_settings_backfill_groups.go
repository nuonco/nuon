package migrations

import (
	"context"
	_ "embed"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"gorm.io/gorm"
)

//go:embed runner_group_settings_backfill_groups.sql
var RunnerGroupSettingsBackfillGroupsSQL string

func (m *Migrations) Migration086RunnerGroupSettingsBackfillGroups(ctx context.Context, db *gorm.DB, cfg *internal.Config) error {
	if res := db.WithContext(ctx).
		Exec(RunnerGroupSettingsBackfillGroupsSQL); res.Error != nil {
		return res.Error
	}

	return nil
}
