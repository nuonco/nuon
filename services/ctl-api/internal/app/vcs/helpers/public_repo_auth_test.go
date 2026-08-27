package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPublicRepoAuthDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE vcs_connections (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			github_install_id TEXT NOT NULL DEFAULT '',
			github_account_id TEXT NOT NULL DEFAULT '',
			github_account_name TEXT NOT NULL DEFAULT '',
			created_by_id TEXT NOT NULL DEFAULT '',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER NOT NULL DEFAULT 0
		)
	`).Error)
	return db
}

func TestResolvePublicRepoGithubClientFallsBackWithoutConnection(t *testing.T) {
	db := setupPublicRepoAuthDB(t)
	h := &Helpers{db: db, l: zap.NewNop()}

	client, authenticated, err := h.ResolvePublicRepoGithubClient(context.Background(), h.l, "org-1", "acme")
	require.NoError(t, err)
	require.False(t, authenticated)
	require.NotNil(t, client)
}

func TestFindOrgVCSConnectionForOwner(t *testing.T) {
	db := setupPublicRepoAuthDB(t)
	require.NoError(t, db.Exec(`
		INSERT INTO vcs_connections (id, org_id, github_account_name, github_install_id, created_at, updated_at)
		VALUES ('conn-1', 'org-1', 'acme', '123', datetime('now'), datetime('now'))
	`).Error)

	h := &Helpers{db: db}
	conn, found, err := h.findOrgVCSConnectionForOwner(context.Background(), "org-1", "acme")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "conn-1", conn.ID)

	_, found, err = h.findOrgVCSConnectionForOwner(context.Background(), "org-1", "other")
	require.NoError(t, err)
	require.False(t, found)
}
