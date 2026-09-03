package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstallPreflightReturnsAllFindings(t *testing.T) {
	db := stackOutdatedTestDB(t)
	for _, statement := range []string{
		`CREATE TABLE components (id TEXT PRIMARY KEY, name TEXT, deleted_at INTEGER DEFAULT 0)`,
		`CREATE TABLE component_builds (id TEXT PRIMARY KEY, status TEXT, status_description TEXT, component_config_connection_id TEXT, deleted_at INTEGER DEFAULT 0)`,
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	seedStackOutdatedTest(t, db, "appcfg-old", "appcfg-new", true)
	seedAppStackConfig(t, db, "stackcfg-old", "appcfg-old", "https://example.com/runner-v1.tf")
	seedAppStackConfig(t, db, "stackcfg-new", "appcfg-new", "https://example.com/runner-v2.tf")
	require.NoError(t, db.Exec(`INSERT INTO components (id, name) VALUES (?, ?)`, "cmp123", "api").Error)

	result, err := (&Activities{db: db}).runInstallPreflightChecks(context.Background(), InstallPreflightRequest{
		FlowID:             "iwf123",
		InstallID:          "inl123",
		CheckStackOutdated: true,
		PlannedComponentBuilds: []PlannedComponentBuild{
			{ComponentID: "cmp123", BuildID: "bld-missing"},
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Findings, 2)
}
