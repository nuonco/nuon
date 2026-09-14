package customermanaged

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLoadPinnedSandboxBuildUsesReleaseOwnership(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE app_sandbox_builds (
		id text PRIMARY KEY,
		org_id text,
		app_id text,
		app_config_id text,
		app_sandbox_config_id text,
		deleted_at integer DEFAULT 0
	)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO app_sandbox_builds
		(id, org_id, app_id, app_config_id, app_sandbox_config_id, deleted_at)
		VALUES (?, ?, ?, ?, ?, 0)`, "build-a", "org-a", "app-a", "config-old", "sandbox-old").Error)

	build, err := LoadPinnedSandboxBuild(context.Background(), db, "org-a", "app-a", "build-a")
	require.NoError(t, err)
	require.Equal(t, "config-old", build.AppConfigID)

	_, err = LoadPinnedSandboxBuild(context.Background(), db, "org-b", "app-a", "build-a")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = LoadPinnedSandboxBuild(context.Background(), db, "org-a", "app-b", "build-a")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
