package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEnqueueAppBranchCreatedIfFirstSkipsWhenNotFirstConfig(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE app_branch_configs (
			id TEXT PRIMARY KEY,
			app_branch_id TEXT NOT NULL,
			deleted_at INTEGER NOT NULL DEFAULT 0
		)
	`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO app_branch_configs (id, app_branch_id) VALUES ('cfg-1', 'branch-1'), ('cfg-2', 'branch-1')
	`).Error)

	h := &Helpers{db: db}
	require.NoError(t, h.EnqueueAppBranchCreatedIfFirst(context.Background(), "branch-1", "cfg-2"))
}

func TestEnqueueAppBranchCreatedIfFirstRequiresIDs(t *testing.T) {
	h := &Helpers{}
	require.Error(t, h.EnqueueAppBranchCreatedIfFirst(context.Background(), "", "cfg-1"))
	require.Error(t, h.EnqueueAppBranchCreatedIfFirst(context.Background(), "branch-1", ""))
}
