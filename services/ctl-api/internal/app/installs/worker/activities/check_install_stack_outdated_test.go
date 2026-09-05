package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/deployerrors"
)

func stackOutdatedTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	for _, statement := range []string{
		`CREATE TABLE installs (id TEXT PRIMARY KEY, app_config_id TEXT, deleted_at INTEGER DEFAULT 0)`,
		`CREATE TABLE install_stack_versions (id TEXT PRIMARY KEY, install_id TEXT, app_config_id TEXT, status BLOB, created_at DATETIME, deleted_at INTEGER DEFAULT 0)`,
		`CREATE TABLE app_stack_configs (id TEXT PRIMARY KEY, app_config_id TEXT, type TEXT, name TEXT, description TEXT, runner_nested_template_url TEXT, vpc_nested_template_url TEXT, deployment_scope TEXT, custom_nested_stacks JSON, deleted_at INTEGER DEFAULT 0)`,
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	return db
}

func seedStackOutdatedTest(t *testing.T, db *gorm.DB, appliedConfig, desiredConfig string, active bool) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO installs (id, app_config_id) VALUES (?, ?)`, "inl123", desiredConfig).Error)
	status := "pending_user"
	if active {
		status = "active"
	}
	require.NoError(t, db.Exec(
		`INSERT INTO install_stack_versions (id, install_id, app_config_id, status, created_at) VALUES (?, ?, ?, ?, datetime('now'))`,
		"isv123", "inl123", appliedConfig, []byte(`{"status":"`+status+`"}`),
	).Error)
}

func seedAppStackConfig(t *testing.T, db *gorm.DB, id, appConfigID, runnerURL string) {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO app_stack_configs (id, app_config_id, type, name, runner_nested_template_url, custom_nested_stacks) VALUES (?, ?, ?, ?, ?, '[]')`,
		id, appConfigID, "gcp-terraform", "acme-stack", runnerURL,
	).Error)
}

func runStackOutdatedCheck(t *testing.T, db *gorm.DB, desiredAppConfigID string) int {
	t.Helper()
	findings, err := (&Activities{db: db}).checkInstallStackOutdated(context.Background(), InstallPreflightRequest{
		FlowID:             "iwf123",
		InstallID:          "inl123",
		DesiredAppConfigID: desiredAppConfigID,
		CheckStackOutdated: true,
	})
	require.NoError(t, err)
	if len(findings) > 0 {
		require.Equal(t, deployerrors.InstallStackOutdatedErrorType, findings[0].Type)
		require.Equal(t, "install_workflows", findings[0].SourceType)
		require.Equal(t, "iwf123", findings[0].SourceID)
	}
	return len(findings)
}

func TestCheckInstallStackOutdated(t *testing.T) {
	t.Run("unchanged content with different app configs", func(t *testing.T) {
		db := stackOutdatedTestDB(t)
		seedStackOutdatedTest(t, db, "appcfg-old", "appcfg-new", true)
		seedAppStackConfig(t, db, "stackcfg-old", "appcfg-old", "https://example.com/runner.tf")
		seedAppStackConfig(t, db, "stackcfg-new", "appcfg-new", "https://example.com/runner.tf")

		require.Zero(t, runStackOutdatedCheck(t, db, ""))
	})

	t.Run("changed content with old active version", func(t *testing.T) {
		db := stackOutdatedTestDB(t)
		seedStackOutdatedTest(t, db, "appcfg-old", "appcfg-new", true)
		seedAppStackConfig(t, db, "stackcfg-old", "appcfg-old", "https://example.com/runner-v1.tf")
		seedAppStackConfig(t, db, "stackcfg-new", "appcfg-new", "https://example.com/runner-v2.tf")

		require.Equal(t, 1, runStackOutdatedCheck(t, db, ""))
	})

	t.Run("pending current version does not replace the active version", func(t *testing.T) {
		db := stackOutdatedTestDB(t)
		seedStackOutdatedTest(t, db, "appcfg-old", "appcfg-new", true)
		require.NoError(t, db.Exec(
			`INSERT INTO install_stack_versions (id, install_id, app_config_id, status, created_at) VALUES (?, ?, ?, ?, datetime('now', '+1 second'))`,
			"isv-new", "inl123", "appcfg-new", []byte(`{"status":"awaiting-user-run"}`),
		).Error)
		seedAppStackConfig(t, db, "stackcfg-old", "appcfg-old", "https://example.com/runner-v1.tf")
		seedAppStackConfig(t, db, "stackcfg-new", "appcfg-new", "https://example.com/runner-v2.tf")

		require.Equal(t, 1, runStackOutdatedCheck(t, db, ""))
	})

	t.Run("propagated current version uses the stack version that was applied", func(t *testing.T) {
		db := stackOutdatedTestDB(t)
		seedStackOutdatedTest(t, db, "appcfg-old", "appcfg-new", true)
		require.NoError(t, db.Exec(
			`INSERT INTO install_stack_versions (id, install_id, app_config_id, status, created_at) VALUES (?, ?, ?, ?, datetime('now', '+1 second'))`,
			"isv-new", "inl123", "appcfg-new", []byte(`{"status":"active","metadata":{"applied_from_version_id":"isv123"}}`),
		).Error)
		seedAppStackConfig(t, db, "stackcfg-old", "appcfg-old", "https://example.com/runner-v1.tf")
		seedAppStackConfig(t, db, "stackcfg-new", "appcfg-new", "https://example.com/runner-v2.tf")

		require.Equal(t, 1, runStackOutdatedCheck(t, db, ""))
	})

	t.Run("desired app config override", func(t *testing.T) {
		db := stackOutdatedTestDB(t)
		seedStackOutdatedTest(t, db, "appcfg-current", "appcfg-current", true)
		seedAppStackConfig(t, db, "stackcfg-current", "appcfg-current", "https://example.com/runner-v1.tf")
		seedAppStackConfig(t, db, "stackcfg-next", "appcfg-next", "https://example.com/runner-v2.tf")

		require.Equal(t, 1, runStackOutdatedCheck(t, db, "appcfg-next"))
	})

	t.Run("newest active version wins", func(t *testing.T) {
		db := stackOutdatedTestDB(t)
		seedStackOutdatedTest(t, db, "appcfg-old", "appcfg-current", true)
		require.NoError(t, db.Exec(
			`INSERT INTO install_stack_versions (id, install_id, app_config_id, status, created_at) VALUES (?, ?, ?, ?, datetime('now', '+1 second'))`,
			"isv-current", "inl123", "appcfg-current", []byte(`{"status":"active"}`),
		).Error)

		require.Zero(t, runStackOutdatedCheck(t, db, ""))
	})

	t.Run("pending version without an active version", func(t *testing.T) {
		db := stackOutdatedTestDB(t)
		seedStackOutdatedTest(t, db, "appcfg-old", "appcfg-new", false)

		require.Equal(t, 1, runStackOutdatedCheck(t, db, ""))
	})

	t.Run("no stack versions", func(t *testing.T) {
		db := stackOutdatedTestDB(t)
		require.NoError(t, db.Exec(`INSERT INTO installs (id, app_config_id) VALUES (?, ?)`, "inl123", "appcfg-new").Error)

		require.Zero(t, runStackOutdatedCheck(t, db, ""))
	})

	t.Run("check disabled", func(t *testing.T) {
		db := stackOutdatedTestDB(t)
		findings, err := (&Activities{db: db}).checkInstallStackOutdated(context.Background(), InstallPreflightRequest{})
		require.NoError(t, err)
		require.Empty(t, findings)
	})
}
