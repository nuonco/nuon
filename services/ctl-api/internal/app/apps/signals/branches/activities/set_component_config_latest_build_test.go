package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSetLatestBuildDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE component_config_connections (
			id TEXT PRIMARY KEY,
			app_config_id TEXT,
			component_id TEXT,
			latest_build_id TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0
		)
	`).Error)

	return db
}

func TestSetComponentConfigLatestBuild(t *testing.T) {
	db := setupSetLatestBuildDB(t)
	require.NoError(t, db.Exec(`
		INSERT INTO component_config_connections (id, app_config_id, component_id, latest_build_id, created_at, updated_at)
		VALUES ('conn-1', 'cfg-1', 'cmp-1', NULL, datetime('now'), datetime('now'))
	`).Error)

	a := &Activities{db: db}
	err := a.SetComponentConfigLatestBuild(context.Background(), &SetComponentConfigLatestBuildInput{
		AppConfigID: "cfg-1",
		ComponentID: "cmp-1",
		BuildID:     "bld-1",
	})
	require.NoError(t, err)

	var latestBuildID string
	require.NoError(t, db.Raw(`SELECT latest_build_id FROM component_config_connections WHERE id = 'conn-1'`).Scan(&latestBuildID).Error)
	require.Equal(t, "bld-1", latestBuildID)

	err = a.SetComponentConfigLatestBuild(context.Background(), &SetComponentConfigLatestBuildInput{
		AppConfigID: "cfg-missing",
		ComponentID: "cmp-1",
		BuildID:     "bld-2",
	})
	require.Error(t, err)
}
