package components

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func setupReusableBuildDB(t *testing.T) *gorm.DB {
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

func TestReusableActiveBuildIDSkipsDifferentChecksumPrev(t *testing.T) {
	db := setupReusableBuildDB(t)
	now := time.Now().UTC()

	require.NoError(t, db.Exec(`
		INSERT INTO component_config_connections (id, component_id, checksum, latest_build_id, created_at, updated_at)
		VALUES
			('ccc-old', 'cmp-1', 'sum-a', 'bld-old', ?, ?),
			('ccc-other', 'cmp-1', 'sum-b', 'bld-other', ?, ?),
			('ccc-new', 'cmp-1', 'sum-a', NULL, ?, ?)
	`, now.Add(-2*time.Hour), now.Add(-2*time.Hour),
		now.Add(-1*time.Hour), now.Add(-1*time.Hour),
		now, now).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO component_builds (id, component_config_connection_id, status, created_at, updated_at)
		VALUES
			('bld-old', 'ccc-old', 'active', ?, ?),
			('bld-other', 'ccc-other', 'active', ?, ?)
	`, now.Add(-2*time.Hour), now.Add(-2*time.Hour),
		now.Add(-1*time.Hour), now.Add(-1*time.Hour)).Error)

	incoming := &app.ComponentConfigConnection{
		ID:       "ccc-new",
		Checksum: "sum-a",
	}

	found, buildID, err := reusableActiveBuildID(context.Background(), db, "cmp-1", incoming)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "bld-old", buildID)
}
