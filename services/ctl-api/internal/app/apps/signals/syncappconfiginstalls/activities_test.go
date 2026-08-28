package syncappconfiginstalls

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetNonBranchManagedInstallIDsRespectsLatestCommandAuthority(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE installs (
		id text primary key, app_id text, deleted_at integer default 0
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE install_management_policy_versions (
		id text primary key, install_id text, version integer, command_authority text
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
		use_for_previews boolean
	)`).Error)

	require.NoError(t, db.Exec(`INSERT INTO installs (id, app_id) VALUES
		('legacy', 'app-a'),
		('managed', 'app-a'),
		('customer', 'app-a'),
		('became-managed', 'app-a'),
		('became-customer', 'app-a'),
		('other-app', 'app-b')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO install_management_policy_versions
		(id, install_id, version, command_authority) VALUES
		('managed-1', 'managed', 1, 'nuon'),
		('customer-1', 'customer', 1, 'customer'),
		('became-managed-1', 'became-managed', 1, 'customer'),
		('became-managed-2', 'became-managed', 2, 'nuon'),
		('became-customer-1', 'became-customer', 1, 'nuon'),
		('became-customer-2', 'became-customer', 2, 'customer')`).Error)

	activities := NewActivities(ActivitiesParams{DB: db})
	result, err := activities.GetNonBranchManagedInstallIDs(context.Background(), &GetNonBranchManagedInstallIDsInput{AppID: "app-a"})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"legacy", "managed", "became-managed"}, result.InstallIDs)
}
