package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration142CloudConnections(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
CREATE TABLE IF NOT EXISTS cloud_connections (
    id text PRIMARY KEY CHECK (char_length(id) = 26),
    created_by_id text NOT NULL REFERENCES accounts(id),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at bigint NOT NULL DEFAULT 0,
    org_id text NOT NULL REFERENCES orgs(id),
    name text NOT NULL,
    platform text NOT NULL,
    target_id text NOT NULL,
    principal text NOT NULL,
    tenant_id text,
    identity_provider text,
    default_region text NOT NULL DEFAULT '',
    auth_mode text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'pending',
    status_message text NOT NULL DEFAULT '',
    last_verified_at timestamptz,
    verification_requested_at timestamptz,
    preset text NOT NULL CHECK (preset IN ('stacks', 'custom'))
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cloud_connections_org_platform_principal_deleted
    ON cloud_connections (org_id, platform, principal, deleted_at);
ALTER TABLE installs ADD COLUMN IF NOT EXISTS cloud_connection_id text REFERENCES cloud_connections(id);
ALTER TABLE aws_accounts DROP COLUMN IF EXISTS aws_account_connection_id;
DROP TABLE IF EXISTS aws_account_connections;
`).Error
}
