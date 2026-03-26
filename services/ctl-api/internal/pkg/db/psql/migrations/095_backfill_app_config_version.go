package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration095BackfillAppConfigVersion(ctx context.Context, db *gorm.DB) error {
	backfill := `
UPDATE app_configs
SET version = v.version
FROM (
    SELECT id, row_number() OVER (PARTITION BY app_id ORDER BY created_at) AS version
    FROM app_configs
) v
WHERE app_configs.id = v.id;
`
	if res := db.WithContext(ctx).Exec(backfill); res.Error != nil {
		return res.Error
	}

	return nil
}
