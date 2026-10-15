package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func (m *Migrations) Migration146InstallStateHeads(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(fmt.Sprintf(`
INSERT INTO install_state_heads (install_id, org_id, revision, current_state_id, stale_at, stale_partials)
SELECT i.id, i.org_id, 1, '', now(), to_jsonb(%s)
FROM installs i
WHERE i.deleted_at = 0
  AND NOT EXISTS (SELECT 1 FROM install_state_heads h WHERE h.install_id = i.id)
ORDER BY i.id
ON CONFLICT (install_id) DO NOTHING`, installStateSQLArray(pkgstate.AllPartials))).Error
}
