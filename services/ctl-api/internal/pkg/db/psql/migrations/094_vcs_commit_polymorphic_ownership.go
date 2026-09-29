package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration094VCSCommitPolymorphicOwnership(ctx context.Context, db *gorm.DB) error {
	addOwnerFields := `
		ALTER TABLE vcs_connection_commits 
			ADD COLUMN IF NOT EXISTS owner_id TEXT,
			ADD COLUMN IF NOT EXISTS owner_type TEXT;
	`
	if res := db.WithContext(ctx).Exec(addOwnerFields); res.Error != nil {
		return res.Error
	}

	makeVCSConnectionIDNullable := `
		ALTER TABLE vcs_connection_commits 
			ALTER COLUMN vcs_connection_id DROP NOT NULL;
	`
	if res := db.WithContext(ctx).Exec(makeVCSConnectionIDNullable); res.Error != nil {
		return res.Error
	}

	createIndex := `
		CREATE INDEX IF NOT EXISTS idx_vcs_commits_owner 
		ON vcs_connection_commits(owner_id, owner_type);
	`
	if res := db.WithContext(ctx).Exec(createIndex); res.Error != nil {
		return res.Error
	}

	addConstraint := `
		ALTER TABLE vcs_connection_commits
			ADD CONSTRAINT owner_id_checker CHECK (owner_id IS NULL OR char_length(owner_id) = 26);
	`
	if res := db.WithContext(ctx).Exec(addConstraint); res.Error != nil {
		return res.Error
	}

	return nil
}
