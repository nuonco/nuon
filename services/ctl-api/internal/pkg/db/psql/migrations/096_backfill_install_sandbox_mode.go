package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration096BackfillInstallSandboxMode(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec(`UPDATE installs
SET sandbox_mode = orgs.sandbox_mode
FROM orgs
WHERE installs.org_id = orgs.id
AND installs.sandbox_mode IS NULL;`); res.Error != nil {
		return res.Error
	}

	return nil
}
