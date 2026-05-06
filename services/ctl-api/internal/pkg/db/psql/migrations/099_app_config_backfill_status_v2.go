package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration099AppConfigBackfillStatusV2 backfills the status_v2 JSONB column
// from the legacy v1 status + status_description fields for all app_configs
// that don't yet have a v2 status set.
func (m *Migrations) Migration099AppConfigBackfillStatusV2(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec(`UPDATE app_configs
SET status_v2 = jsonb_build_object(
    'status', status,
    'status_human_description', COALESCE(status_description, ''),
    'created_by_id', COALESCE(created_by_id, ''),
    'created_at_ts', EXTRACT(EPOCH FROM updated_at)::bigint,
    'metadata', jsonb_build_object('migrated', true),
    'history', '[]'::jsonb
)
WHERE deleted_at = 0
AND (status_v2 IS NULL OR status_v2 = '{}'::jsonb OR status_v2 = 'null'::jsonb);`); res.Error != nil {
		return res.Error
	}

	return nil
}
