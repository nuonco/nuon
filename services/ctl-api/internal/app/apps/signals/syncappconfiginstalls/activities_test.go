package syncappconfiginstalls

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetNonBranchManagedInstallIDsRespectsApprovalAuthority(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE installs (
		id text primary key, app_id text, deleted_at integer default 0
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE install_operating_models (
		id text primary key, install_id text unique, approval_authority text
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE app_branches (
		id text primary key, app_id text, deleted_at integer default 0
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE app_branch_configs (
		id text primary key, app_branch_id text
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE app_branch_install_groups (
		id text primary key, created_by_id text, created_at datetime, updated_at datetime,
		deleted_at integer default 0, org_id text, app_branch_config_id text, name text,
		"order" integer, install_ids text, label_selector text, max_parallel integer,
		use_for_previews boolean, all_installs boolean
	)`).Error)

	require.NoError(t, db.Exec(`INSERT INTO installs (id, app_id) VALUES
		('legacy', 'app-a'),
		('managed', 'app-a'),
		('customer', 'app-a'),
		('other-app', 'app-b')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO install_operating_models
		(id, install_id, approval_authority) VALUES
		('managed', 'managed', 'vendor'),
		('customer', 'customer', 'customer')`).Error)

	activities := NewActivities(ActivitiesParams{DB: db})
	result, err := activities.GetNonBranchManagedInstallIDs(context.Background(), &GetNonBranchManagedInstallIDsInput{AppID: "app-a"})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"legacy", "managed"}, result.InstallIDs)
}
