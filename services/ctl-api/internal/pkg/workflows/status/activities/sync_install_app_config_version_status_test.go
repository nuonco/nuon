package statusactivities

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func setupIACVSyncDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE install_app_config_versions (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL DEFAULT '',
			install_id TEXT NOT NULL DEFAULT '',
			old_app_config_id TEXT NOT NULL DEFAULT '',
			new_app_config_id TEXT NOT NULL DEFAULT '',
			workflow_id TEXT,
			app_release_id TEXT,
			policy_version_id TEXT,
			app_branch_run_id TEXT,
			install_group_id TEXT,
			created_by_id TEXT NOT NULL DEFAULT '',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			status BLOB,
			diff TEXT,
			metadata TEXT
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE install_release_deployments (
			id TEXT PRIMARY KEY, created_at DATETIME, org_id TEXT, install_id TEXT,
			release_id TEXT, package_id TEXT, previous_release_id TEXT,
			install_app_config_version_id TEXT, policy_version_id TEXT, method TEXT,
			actor TEXT, executor TEXT, operation_id TEXT, plan_digest TEXT,
			result_directive TEXT, status TEXT, started_at DATETIME, finished_at DATETIME
		)
	`).Error)
	return db
}

func insertIACV(t *testing.T, db *gorm.DB, id, installID, workflowID string, status app.CompositeStatus) {
	t.Helper()
	statusJSON, err := json.Marshal(status)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		INSERT INTO install_app_config_versions (id, install_id, workflow_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
	`, id, installID, workflowID, statusJSON).Error)
}

func readIACVStatus(t *testing.T, db *gorm.DB, id string) app.CompositeStatus {
	t.Helper()
	var raw string
	require.NoError(t, db.Raw(`SELECT status FROM install_app_config_versions WHERE id = ?`, id).Scan(&raw).Error)
	var status app.CompositeStatus
	require.NoError(t, json.Unmarshal([]byte(raw), &status))
	return status
}

func TestSyncInstallAppConfigVersionFromFlowStatus(t *testing.T) {
	db := setupIACVSyncDB(t)
	wfID := "wf-linked"
	insertIACV(t, db, "iacv-1", "inst-1", wfID, app.CompositeStatus{Status: app.StatusInProgress})
	insertIACV(t, db, "iacv-other", "inst-2", "wf-other", app.CompositeStatus{Status: app.StatusInProgress})

	a := &Activities{db: db, l: zap.NewNop()}
	a.syncInstallAppConfigVersionFromFlowStatus(context.Background(), app.Workflow{ID: wfID}, app.CompositeStatus{
		Status:                 app.StatusError,
		StatusHumanDescription: "workflow failed, awaiting retry",
	})

	linked := readIACVStatus(t, db, "iacv-1")
	require.Equal(t, app.StatusError, linked.Status)
	require.Equal(t, "workflow failed, awaiting retry", linked.StatusHumanDescription)

	other := readIACVStatus(t, db, "iacv-other")
	require.Equal(t, app.StatusInProgress, other.Status)
}

func TestSyncInstallAppConfigVersionFromFlowStatusNoMatch(t *testing.T) {
	db := setupIACVSyncDB(t)
	a := &Activities{db: db, l: zap.NewNop()}

	a.syncInstallAppConfigVersionFromFlowStatus(context.Background(), app.Workflow{ID: "wf-missing"}, app.CompositeStatus{
		Status: app.StatusError,
	})
}

func TestSyncInstallAppConfigVersionRecordsOnlyTerminalReleaseFailure(t *testing.T) {
	db := setupIACVSyncDB(t)
	releaseID := "release-1"
	policyID := "policy-1"
	require.NoError(t, db.Exec(`
		INSERT INTO install_app_config_versions
			(id, org_id, install_id, workflow_id, app_release_id, policy_version_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`, "iacv-release", "org-1", "install-1", "workflow-1", releaseID, policyID, `{}`).Error)
	a := &Activities{db: db, l: zap.NewNop()}
	workflow := app.Workflow{ID: "workflow-1", CreatedAt: time.Now()}

	a.syncInstallAppConfigVersionFromFlowStatus(context.Background(), workflow, app.CompositeStatus{
		Status: app.StatusError, Metadata: map[string]any{"awaiting_retry": true},
	})
	var count int64
	require.NoError(t, db.Model(&app.InstallReleaseDeployment{}).Count(&count).Error)
	require.Zero(t, count)

	a.syncInstallAppConfigVersionFromFlowStatus(context.Background(), workflow, app.CompositeStatus{
		Status: app.StatusError, Metadata: map[string]any{"stopped": true},
	})
	var deployment app.InstallReleaseDeployment
	require.NoError(t, db.First(&deployment).Error)
	require.Equal(t, releaseID, deployment.ReleaseID)
	require.Equal(t, app.InstallDeploymentStatusFailed, deployment.Status)
	require.Equal(t, "failed", deployment.ResultDirective)
}
