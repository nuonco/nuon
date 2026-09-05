package activities

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/deployerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func componentBuildPreflightTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	for _, statement := range []string{
		`CREATE TABLE components (id TEXT PRIMARY KEY, name TEXT, deleted_at INTEGER DEFAULT 0)`,
		`CREATE TABLE component_config_connections (id TEXT PRIMARY KEY, component_id TEXT, checksum TEXT, latest_build_id TEXT, deleted_at INTEGER DEFAULT 0)`,
		`CREATE TABLE component_builds (id TEXT PRIMARY KEY, status TEXT, status_description TEXT, component_config_connection_id TEXT, deleted_at INTEGER DEFAULT 0)`,
		`CREATE TABLE install_components (id TEXT PRIMARY KEY, install_id TEXT, component_id TEXT, deleted_at INTEGER DEFAULT 0)`,
		`CREATE TABLE install_deploys (id TEXT PRIMARY KEY, component_build_id TEXT, install_component_id TEXT, deleted_at INTEGER DEFAULT 0)`,
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	require.NoError(t, db.Exec(`INSERT INTO components (id, name) VALUES (?, ?)`, "cmp123", "api").Error)
	return db
}

func componentBuildPreflightRequest(planned ...PlannedComponentBuild) InstallPreflightRequest {
	return InstallPreflightRequest{
		FlowID:                 "iwf123",
		InstallID:              "inl123",
		PlannedComponentBuilds: planned,
	}
}

func TestCheckComponentBuildsAllowsUsableAndInProgressBuilds(t *testing.T) {
	for _, status := range []app.ComponentBuildStatus{
		app.ComponentBuildStatusActive,
		app.ComponentBuildStatusQueued,
		app.ComponentBuildStatusPlanning,
		app.ComponentBuildStatusBuilding,
	} {
		t.Run(string(status), func(t *testing.T) {
			db := componentBuildPreflightTestDB(t)
			require.NoError(t, db.Exec(
				`INSERT INTO component_builds (id, status) VALUES (?, ?)`, "bld123", status,
			).Error)

			findings, err := (&Activities{db: db}).checkComponentBuilds(context.Background(), componentBuildPreflightRequest(
				PlannedComponentBuild{ComponentID: "cmp123", BuildID: "bld123", WaitForBuild: true},
			))
			require.NoError(t, err)
			require.Empty(t, findings)
		})
	}
}

func TestCheckComponentBuildsRequiresActiveBuildWhenStepDoesNotWait(t *testing.T) {
	db := componentBuildPreflightTestDB(t)
	require.NoError(t, db.Exec(
		`INSERT INTO component_builds (id, status) VALUES (?, ?)`, "bld123", app.ComponentBuildStatusBuilding,
	).Error)

	findings, err := (&Activities{db: db}).checkComponentBuilds(context.Background(), componentBuildPreflightRequest(
		PlannedComponentBuild{ComponentID: "cmp123", BuildID: "bld123"},
	))
	require.NoError(t, err)
	require.Len(t, findings, 1)
}

func TestCheckComponentBuildsReportsMissingBuild(t *testing.T) {
	db := componentBuildPreflightTestDB(t)
	findings, err := (&Activities{db: db}).checkComponentBuilds(context.Background(), componentBuildPreflightRequest(
		PlannedComponentBuild{ComponentID: "cmp123", BuildID: "bld-missing"},
	))
	require.NoError(t, err)
	require.Len(t, findings, 1)
	require.Equal(t, deployerrors.ComponentBuildUnavailableErrorType, findings[0].Type)
	require.Equal(t, compositeerrors.SeverityError, findings[0].Severity)
	require.Equal(t, "install_workflows", findings[0].SourceType)
	require.Equal(t, "iwf123", findings[0].SourceID)

	var details deployerrors.ComponentBuildUnavailableError
	require.NoError(t, json.Unmarshal(findings[0].Data, &details))
	require.Equal(t, deployerrors.ComponentBuildUnavailableReasonMissing, details.Reason)
	require.Equal(t, "api", details.ComponentName)
}

func TestCheckComponentBuildsReportsMissingBuildForConfigConnection(t *testing.T) {
	db := componentBuildPreflightTestDB(t)
	require.NoError(t, db.Exec(
		`INSERT INTO component_config_connections (id, component_id, checksum) VALUES (?, ?, '')`, "ccc123", "cmp123",
	).Error)

	findings, err := (&Activities{db: db}).checkComponentBuilds(context.Background(), componentBuildPreflightRequest(
		PlannedComponentBuild{ComponentID: "cmp123", ComponentConfigConnectionID: "ccc123"},
	))
	require.NoError(t, err)
	require.Len(t, findings, 1)
}

func TestCheckComponentBuildsReportsFailedBuild(t *testing.T) {
	for _, status := range []app.ComponentBuildStatus{
		app.ComponentBuildStatusError,
		app.ComponentBuildStatusPolicyFailed,
	} {
		t.Run(string(status), func(t *testing.T) {
			db := componentBuildPreflightTestDB(t)
			require.NoError(t, db.Exec(
				`INSERT INTO component_builds (id, status, status_description) VALUES (?, ?, ?)`, "bld123", status, "build failed",
			).Error)

			findings, err := (&Activities{db: db}).checkComponentBuilds(context.Background(), componentBuildPreflightRequest(
				PlannedComponentBuild{ComponentID: "cmp123", BuildID: "bld123"},
			))
			require.NoError(t, err)
			require.Len(t, findings, 1)
			require.Equal(t, "component_builds", findings[0].SourceType)
			require.Equal(t, "bld123", findings[0].SourceID)

			var details deployerrors.ComponentBuildUnavailableError
			require.NoError(t, json.Unmarshal(findings[0].Data, &details))
			require.Equal(t, deployerrors.ComponentBuildUnavailableReasonFailed, details.Reason)
			require.Equal(t, string(status), details.BuildStatus)
			require.Equal(t, "build failed", details.BuildStatusDescription)
		})
	}
}

func TestCheckComponentBuildsUsesManualDeployBuild(t *testing.T) {
	db := componentBuildPreflightTestDB(t)
	require.NoError(t, db.Exec(
		`INSERT INTO component_builds (id, status, status_description) VALUES (?, ?, ?)`, "bld123", app.ComponentBuildStatusError, "build failed",
	).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO install_components (id, install_id, component_id) VALUES (?, ?, ?)`, "inc123", "inl123", "cmp123",
	).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO install_deploys (id, component_build_id, install_component_id) VALUES (?, ?, ?)`, "dep123", "bld123", "inc123",
	).Error)

	findings, err := (&Activities{db: db}).checkComponentBuilds(context.Background(), componentBuildPreflightRequest(
		PlannedComponentBuild{DeployID: "dep123"},
	))
	require.NoError(t, err)
	require.Len(t, findings, 1)
	require.Equal(t, "bld123", findings[0].SourceID)
}

func TestCheckComponentBuildsReturnsAllFindings(t *testing.T) {
	db := componentBuildPreflightTestDB(t)
	require.NoError(t, db.Exec(`INSERT INTO components (id, name) VALUES (?, ?)`, "cmp456", "worker").Error)

	findings, err := (&Activities{db: db}).checkComponentBuilds(context.Background(), componentBuildPreflightRequest(
		PlannedComponentBuild{ComponentID: "cmp123", BuildID: "bld123"},
		PlannedComponentBuild{ComponentID: "cmp456", BuildID: "bld456"},
	))
	require.NoError(t, err)
	require.Len(t, findings, 2)
}

func TestComponentConfigConnectionID(t *testing.T) {
	appConfig := &app.AppConfig{ComponentConfigConnections: []app.ComponentConfigConnection{
		{ID: "ccc123", ComponentID: "cmp123"},
	}}
	require.Equal(t, "ccc123", componentConfigConnectionID(appConfig, "cmp123"))
	require.Empty(t, componentConfigConnectionID(appConfig, "cmp456"))
}
