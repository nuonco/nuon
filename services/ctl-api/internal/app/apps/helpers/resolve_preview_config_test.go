package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func setupPreviewInstallCandidatesDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE app_branches (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			deleted_at INTEGER NOT NULL DEFAULT 0
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE installs (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL DEFAULT '',
			app_id TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			app_branch_id TEXT,
			deleted_at INTEGER NOT NULL DEFAULT 0
		)
	`).Error)

	return db
}

func insertPreviewInstall(t *testing.T, db *gorm.DB, id, appID, name string, branchID *string) {
	t.Helper()
	require.NoError(t, db.Exec(`
		INSERT INTO installs (id, app_id, name, app_branch_id)
		VALUES (?, ?, ?, ?)
	`, id, appID, name, branchID).Error)
}

func TestListPreviewInstallCandidates_allAppInstalls(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupPreviewInstallCandidatesDB(t)
	h := &Helpers{db: db}

	appID := "app-1"
	branchID := "branch-1"
	otherBranchID := "branch-2"

	insertPreviewInstall(t, db, "inst-a", appID, "alpha", &branchID)
	insertPreviewInstall(t, db, "inst-b", appID, "beta", nil)
	insertPreviewInstall(t, db, "inst-c", appID, "gamma", &otherBranchID)
	insertPreviewInstall(t, db, "inst-other-app", "app-2", "other", nil)

	installs, err := h.ListPreviewInstallCandidates(ctx, appID, branchID, app.DefaultAppBranchPreviewConfig())
	require.NoError(t, err)
	require.Len(t, installs, 3)

	ids := []string{installs[0].ID, installs[1].ID, installs[2].ID}
	require.ElementsMatch(t, []string{"inst-a", "inst-b", "inst-c"}, ids)
}

func TestBuildAppBranchRunPreview_overrideInstallWithoutBranchDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupPreviewInstallCandidatesDB(t)
	h := &Helpers{db: db}

	appID := "app-1"
	installID := "inst-b"
	insertPreviewInstall(t, db, installID, appID, "beta", nil)

	overrideInstall := installID
	prNumber := 32
	preview, err := h.BuildAppBranchRunPreview(ctx, appID, &app.AppBranchConfig{}, &PreviewRunInput{
		Source:   app.AppBranchRunPreviewSourcePR,
		PRNumber: &prNumber,
		Override: &app.AppBranchPreviewOverride{
			InstallID: &overrideInstall,
		},
	})
	require.NoError(t, err)
	require.Equal(t, installID, preview.InstallID)
	require.Equal(t, "beta", preview.InstallName)
	require.Equal(t, app.AppBranchRunPreviewModePlanOnly, preview.Mode)
}
