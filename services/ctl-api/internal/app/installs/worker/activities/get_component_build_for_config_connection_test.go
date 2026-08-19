package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBuildForCCCDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE component_config_connections (
			id TEXT PRIMARY KEY,
			component_id TEXT,
			checksum TEXT,
			latest_build_id TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE VIEW component_config_connections_view_v1 AS
		SELECT * FROM component_config_connections
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE component_builds (
			id TEXT PRIMARY KEY,
			component_config_connection_id TEXT,
			status TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0
		)
	`).Error)

	return db
}

func TestGetComponentBuildForConfigConnection(t *testing.T) {
	const componentID = "cmp-1"

	insertConn := func(t *testing.T, db *gorm.DB, id, checksum, latestBuildID string) {
		t.Helper()
		require.NoError(t, db.Exec(`
			INSERT INTO component_config_connections (id, component_id, checksum, latest_build_id, created_at, updated_at)
			VALUES (?, ?, ?, nullif(?, ''), datetime('now'), datetime('now'))
		`, id, componentID, checksum, latestBuildID).Error)
	}
	insertBuild := func(t *testing.T, db *gorm.DB, id, connID, status, createdAt string) {
		t.Helper()
		require.NoError(t, db.Exec(`
			INSERT INTO component_builds (id, component_config_connection_id, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, datetime('now'))
		`, id, connID, status, createdAt).Error)
	}

	t.Run("returns pinned active build", func(t *testing.T) {
		db := setupBuildForCCCDB(t)
		insertConn(t, db, "conn-1", "sum", "bld-1")
		insertBuild(t, db, "bld-1", "conn-1", "active", "2026-01-01 00:00:00")

		a := &Activities{db: db}
		out, err := a.GetComponentBuildForConfigConnection(context.Background(), GetComponentBuildForConfigConnectionRequest{
			ComponentConfigConnectionID: "conn-1",
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Equal(t, "bld-1", out.ID)
	})

	t.Run("pinned build not active falls back to checksum match", func(t *testing.T) {
		db := setupBuildForCCCDB(t)
		insertConn(t, db, "conn-old", "sum", "")
		insertBuild(t, db, "bld-old", "conn-old", "active", "2026-01-01 00:00:00")
		insertConn(t, db, "conn-new", "sum", "bld-queued")
		insertBuild(t, db, "bld-queued", "conn-new", "queued", "2026-01-02 00:00:00")

		a := &Activities{db: db}
		out, err := a.GetComponentBuildForConfigConnection(context.Background(), GetComponentBuildForConfigConnectionRequest{
			ComponentConfigConnectionID: "conn-new",
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Equal(t, "bld-old", out.ID)
	})

	t.Run("checksum fallback ignores other checksums", func(t *testing.T) {
		db := setupBuildForCCCDB(t)
		insertConn(t, db, "conn-other", "other-sum", "")
		insertBuild(t, db, "bld-other", "conn-other", "active", "2026-01-01 00:00:00")
		insertConn(t, db, "conn-1", "sum", "")

		a := &Activities{db: db}
		out, err := a.GetComponentBuildForConfigConnection(context.Background(), GetComponentBuildForConfigConnectionRequest{
			ComponentConfigConnectionID: "conn-1",
		})
		require.NoError(t, err)
		require.Nil(t, out)
	})

	t.Run("empty checksum without pinned build returns nil", func(t *testing.T) {
		db := setupBuildForCCCDB(t)
		insertConn(t, db, "conn-1", "", "")

		a := &Activities{db: db}
		out, err := a.GetComponentBuildForConfigConnection(context.Background(), GetComponentBuildForConfigConnectionRequest{
			ComponentConfigConnectionID: "conn-1",
		})
		require.NoError(t, err)
		require.Nil(t, out)
	})

	t.Run("checksum fallback picks newest active", func(t *testing.T) {
		db := setupBuildForCCCDB(t)
		insertConn(t, db, "conn-a", "sum", "")
		insertBuild(t, db, "bld-a", "conn-a", "active", "2026-01-01 00:00:00")
		insertConn(t, db, "conn-b", "sum", "")
		insertBuild(t, db, "bld-b", "conn-b", "active", "2026-01-02 00:00:00")

		a := &Activities{db: db}
		out, err := a.GetComponentBuildForConfigConnection(context.Background(), GetComponentBuildForConfigConnectionRequest{
			ComponentConfigConnectionID: "conn-a",
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Equal(t, "bld-b", out.ID)
	})
}
