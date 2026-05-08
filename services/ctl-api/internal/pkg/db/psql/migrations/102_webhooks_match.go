package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration102WebhooksMatch extends the webhooks unique index to include the
// new match_canonical column so a single (org, url) can be subscribed
// multiple times with different routing predicates (e.g., once for "all
// installs" and again for "components in env=prod"). The actual `match` and
// `match_canonical` columns, plus the new composite index
// `idx_webhooks_org_url_match`, are added by GORM AutoMigrate (the
// `gorm-migrations` step in plugins/migrations/exec.go) from the
// `app.Webhook` struct tags. AutoMigrate runs BEFORE this custom migration,
// but it does NOT drop indexes that have been removed from struct tags —
// so the legacy `idx_webhooks_org_url` lingers and must be dropped here.
//
// On an existing DB the effective sequence is:
//  1. AutoMigrate adds `match` (jsonb, nullable) and
//     `match_canonical` (text not null, defaulting to the empty string).
//     Existing rows are backfilled to the empty string by that default.
//  2. AutoMigrate creates `idx_webhooks_org_url_match` on
//     (org_id, webhook_url, match_canonical, deleted_at). Because every
//     existing row has an empty match_canonical, uniqueness reduces to
//     (org_id, webhook_url, deleted_at), already enforced by the legacy
//     index, so the new index builds without conflicts.
//  3. This migration drops the now-redundant legacy `idx_webhooks_org_url`.
//
// Idempotent via DROP INDEX IF EXISTS so re-runs on already-migrated
// environments are no-ops.
func (m *Migrations) Migration102WebhooksMatch(ctx context.Context, db *gorm.DB) error {
	stmts := []string{
		`DROP INDEX IF EXISTS idx_webhooks_org_url;`,
	}
	for _, qry := range stmts {
		if res := db.WithContext(ctx).Exec(qry); res.Error != nil {
			return res.Error
		}
	}
	return nil
}
