package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAdoptQueuedBuildDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE component_config_connections (
			id TEXT PRIMARY KEY,
			app_config_id TEXT,
			component_id TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE component_builds (
			id TEXT PRIMARY KEY,
			component_config_connection_id TEXT,
			status TEXT,
			app_branch_run_id TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0
		)
	`).Error)

	return db
}

func TestAdoptQueuedComponentBuild(t *testing.T) {
	const componentID = "cmp-1"

	insertConn := func(t *testing.T, db *gorm.DB, id, appConfigID string) {
		t.Helper()
		require.NoError(t, db.Exec(`
			INSERT INTO component_config_connections (id, app_config_id, component_id, created_at, updated_at)
			VALUES (?, ?, ?, datetime('now'), datetime('now'))
		`, id, appConfigID, componentID).Error)
	}
	insertBuild := func(t *testing.T, db *gorm.DB, id, connID, status string) {
		t.Helper()
		require.NoError(t, db.Exec(`
			INSERT INTO component_builds (id, component_config_connection_id, status, created_at, updated_at)
			VALUES (?, ?, ?, datetime('now'), datetime('now'))
		`, id, connID, status).Error)
	}

	t.Run("adopts queued build and stamps branch run", func(t *testing.T) {
		db := setupAdoptQueuedBuildDB(t)
		insertConn(t, db, "conn-1", "cfg-1")
		insertBuild(t, db, "bld-1", "conn-1", "queued")

		a := &Activities{db: db}
		out, err := a.AdoptQueuedComponentBuild(context.Background(), AdoptQueuedComponentBuildRequest{
			ComponentID:    componentID,
			AppConfigID:    "cfg-1",
			AppBranchRunID: "run-1",
		})
		require.NoError(t, err)
		require.Equal(t, "bld-1", out.BuildID)

		var runID string
		require.NoError(t, db.Raw(`SELECT app_branch_run_id FROM component_builds WHERE id = 'bld-1'`).Scan(&runID).Error)
		require.Equal(t, "run-1", runID)
	})

	t.Run("ignores non-queued builds", func(t *testing.T) {
		db := setupAdoptQueuedBuildDB(t)
		insertConn(t, db, "conn-1", "cfg-1")
		insertBuild(t, db, "bld-1", "conn-1", "active")

		a := &Activities{db: db}
		out, err := a.AdoptQueuedComponentBuild(context.Background(), AdoptQueuedComponentBuildRequest{
			ComponentID: componentID,
			AppConfigID: "cfg-1",
		})
		require.NoError(t, err)
		require.Empty(t, out.BuildID)
	})

	t.Run("ignores queued builds on other app configs", func(t *testing.T) {
		db := setupAdoptQueuedBuildDB(t)
		insertConn(t, db, "conn-old", "cfg-old")
		insertBuild(t, db, "bld-old", "conn-old", "queued")

		a := &Activities{db: db}
		out, err := a.AdoptQueuedComponentBuild(context.Background(), AdoptQueuedComponentBuildRequest{
			ComponentID: componentID,
			AppConfigID: "cfg-new",
		})
		require.NoError(t, err)
		require.Empty(t, out.BuildID)
	})

	t.Run("no config connection returns empty", func(t *testing.T) {
		db := setupAdoptQueuedBuildDB(t)

		a := &Activities{db: db}
		out, err := a.AdoptQueuedComponentBuild(context.Background(), AdoptQueuedComponentBuildRequest{
			ComponentID: componentID,
			AppConfigID: "cfg-1",
		})
		require.NoError(t, err)
		require.Empty(t, out.BuildID)
	})
}
