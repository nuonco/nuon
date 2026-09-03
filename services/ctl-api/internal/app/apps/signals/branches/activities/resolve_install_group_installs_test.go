package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

func TestResolveInstallGroupInstallsRespectsOperatingModel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE installs (
		id text primary key, app_id text, app_branch_id text, labels text, deleted_at integer default 0
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE install_operating_models (
		id text primary key, install_id text unique, connectivity text,
		release_selection text, approval_authority text
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE app_branches (
		id text primary key, app_id text, deleted_at integer default 0
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE app_branch_configs (
		id text primary key, app_branch_id text, created_at datetime
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE app_branch_install_groups (
		id text primary key, app_branch_config_id text, label_selector text, deleted_at integer default 0
	)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO installs (id, app_id) VALUES
		('legacy', 'app-a'),
		('vendor', 'app-a'),
		('customer', 'app-a'),
		('offline', 'app-a')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO install_operating_models
	(id, install_id, connectivity, release_selection, approval_authority) VALUES
		('vendor', 'vendor', 'connected', 'vendor_proposed', 'vendor'),
		('customer', 'customer', 'connected', 'vendor_proposed', 'customer'),
		('offline', 'offline', 'offline', 'customer', 'customer')`).Error)

	activity := &Activities{
		db:      db,
		helpers: helpers.New(helpers.Params{DB: db}),
		l:       zap.NewNop(),
	}

	t.Run("pinned install IDs", func(t *testing.T) {
		result, err := activity.ResolveInstallGroupInstalls(context.Background(), &ResolveInstallGroupInstallsInput{
			GroupID:    "pinned",
			InstallIDs: []string{"customer", "offline", "vendor", "legacy"},
		})
		require.NoError(t, err)
		require.Equal(t, []string{"customer", "vendor", "legacy"}, result.InstallIDs)
	})

	t.Run("all installs", func(t *testing.T) {
		result, err := activity.ResolveInstallGroupInstalls(context.Background(), &ResolveInstallGroupInstallsInput{
			AppID:       "app-a",
			AppBranchID: "branch-a",
			GroupID:     "all",
			AllInstalls: true,
		})
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"legacy", "vendor", "customer"}, result.InstallIDs)
	})
}
