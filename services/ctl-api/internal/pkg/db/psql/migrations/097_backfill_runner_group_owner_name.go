package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration097BackfillRunnerGroupOwnerName(ctx context.Context, db *gorm.DB) error {
	addColumn := `ALTER TABLE runner_groups ADD COLUMN IF NOT EXISTS owner_name TEXT;`
	if res := db.WithContext(ctx).Exec(addColumn); res.Error != nil {
		return res.Error
	}

	backfillInstalls := `UPDATE runner_groups
SET owner_name = installs.name
FROM installs
WHERE runner_groups.owner_type = 'installs' AND runner_groups.owner_id = installs.id;`
	if res := db.WithContext(ctx).Exec(backfillInstalls); res.Error != nil {
		return res.Error
	}

	backfillOrgs := `UPDATE runner_groups
SET owner_name = orgs.name
FROM orgs
WHERE runner_groups.owner_type = 'orgs' AND runner_groups.owner_id = orgs.id;`
	if res := db.WithContext(ctx).Exec(backfillOrgs); res.Error != nil {
		return res.Error
	}

	return nil
}
