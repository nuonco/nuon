package migrations

import (
	"context"
)

func (m *Migrations) migration054DropSandboxReleases(ctx context.Context) error {
	sql := `
ALTER TABLE app_sandbox_configs DROP COLUMN IF EXISTS sandbox_release_id;
ALTER TABLE apps DROP COLUMN IF EXISTS sandbox_release_id;
DROP TABLE IF EXISTS sandbox_releases;
DROP TABLE IF EXISTS sandboxes;
`

	if res := m.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
