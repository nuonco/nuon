package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCheckBuildNeededDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE component_config_connections (
			id TEXT PRIMARY KEY,
			app_config_id TEXT,
			component_id TEXT,
			checksum TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE external_image_component_configs (
			id TEXT PRIMARY KEY,
			component_config_connection_id TEXT,
			update_policy TEXT,
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
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0
		)
	`).Error)

	return db
}

func insertConnection(t *testing.T, db *gorm.DB, id, appConfigID, componentID, checksum string) {
	t.Helper()
	require.NoError(t, db.Exec(`
		INSERT INTO component_config_connections (id, app_config_id, component_id, checksum, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
	`, id, appConfigID, componentID, checksum).Error)
}

func TestCheckBuildNeeded(t *testing.T) {
	const componentID = "cmp-1"

	t.Run("unchanged checksum with active build reuses build", func(t *testing.T) {
		db := setupCheckBuildNeededDB(t)
		insertConnection(t, db, "conn-old", "cfg-old", componentID, "sum")
		insertConnection(t, db, "conn-new", "cfg-new", componentID, "sum")
		require.NoError(t, db.Exec(`
			INSERT INTO component_builds (id, component_config_connection_id, status, created_at, updated_at)
			VALUES ('bld-1', 'conn-old', 'active', datetime('now'), datetime('now'))
		`).Error)

		a := &Activities{db: db}
		out, err := a.CheckBuildNeeded(context.Background(), &CheckBuildNeededInput{
			ComponentID:    componentID,
			NewAppConfigID: "cfg-new",
			OldAppConfigID: "cfg-old",
		})
		require.NoError(t, err)
		require.False(t, out.NeedsBuild)
		require.Equal(t, "bld-1", out.ExistingBuildID)
	})

	t.Run("update_policy always rebuilds even when checksum unchanged", func(t *testing.T) {
		db := setupCheckBuildNeededDB(t)
		insertConnection(t, db, "conn-old", "cfg-old", componentID, "sum")
		insertConnection(t, db, "conn-new", "cfg-new", componentID, "sum")
		require.NoError(t, db.Exec(`
			INSERT INTO external_image_component_configs (id, component_config_connection_id, update_policy, created_at, updated_at)
			VALUES ('img-1', 'conn-new', '~1.10.0', datetime('now'), datetime('now'))
		`).Error)
		require.NoError(t, db.Exec(`
			INSERT INTO component_builds (id, component_config_connection_id, status, created_at, updated_at)
			VALUES ('bld-1', 'conn-old', 'active', datetime('now'), datetime('now'))
		`).Error)

		a := &Activities{db: db}
		out, err := a.CheckBuildNeeded(context.Background(), &CheckBuildNeededInput{
			ComponentID:    componentID,
			NewAppConfigID: "cfg-new",
			OldAppConfigID: "cfg-old",
		})
		require.NoError(t, err)
		require.True(t, out.NeedsBuild)
	})

	t.Run("external image without update_policy still reuses build", func(t *testing.T) {
		db := setupCheckBuildNeededDB(t)
		insertConnection(t, db, "conn-old", "cfg-old", componentID, "sum")
		insertConnection(t, db, "conn-new", "cfg-new", componentID, "sum")
		require.NoError(t, db.Exec(`
			INSERT INTO external_image_component_configs (id, component_config_connection_id, update_policy, created_at, updated_at)
			VALUES ('img-1', 'conn-new', '', datetime('now'), datetime('now'))
		`).Error)
		require.NoError(t, db.Exec(`
			INSERT INTO component_builds (id, component_config_connection_id, status, created_at, updated_at)
			VALUES ('bld-1', 'conn-old', 'active', datetime('now'), datetime('now'))
		`).Error)

		a := &Activities{db: db}
		out, err := a.CheckBuildNeeded(context.Background(), &CheckBuildNeededInput{
			ComponentID:    componentID,
			NewAppConfigID: "cfg-new",
			OldAppConfigID: "cfg-old",
		})
		require.NoError(t, err)
		require.False(t, out.NeedsBuild)
		require.Equal(t, "bld-1", out.ExistingBuildID)
	})

	t.Run("no previous config always builds", func(t *testing.T) {
		db := setupCheckBuildNeededDB(t)

		a := &Activities{db: db}
		out, err := a.CheckBuildNeeded(context.Background(), &CheckBuildNeededInput{
			ComponentID:    componentID,
			NewAppConfigID: "cfg-new",
		})
		require.NoError(t, err)
		require.True(t, out.NeedsBuild)
	})
}
