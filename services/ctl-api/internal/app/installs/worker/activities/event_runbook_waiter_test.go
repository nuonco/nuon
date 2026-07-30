package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func setupWaiterDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TABLE event_runbook_waiters (
		id TEXT PRIMARY KEY, created_at DATETIME, updated_at DATETIME,
		org_id TEXT, app_id TEXT, install_id TEXT, workflow_id TEXT, workflow_step_id TEXT,
		queue_signal_id TEXT, trigger_id TEXT, event_types TEXT, filters TEXT,
		status TEXT, matched_event_id TEXT, activated_at DATETIME,
		matched_at DATETIME, notified_at DATETIME, cancelled_at DATETIME, expired_at DATETIME
	)`).Error)
	return db
}

func insertActiveWaiter(t *testing.T, db *gorm.DB, id, installID, workflowStepID string) {
	t.Helper()
	require.NoError(t, db.Exec(`
		INSERT INTO event_runbook_waiters (
			id, created_at, updated_at, org_id, app_id, install_id, workflow_id, workflow_step_id,
			queue_signal_id, trigger_id, status, activated_at
		) VALUES (?, datetime('now'), datetime('now'), 'org-a', 'app-1', ?, 'wf-1', ?,
			'qs-1', 'trg-1', 'active', datetime('now'))
	`, id, installID, workflowStepID).Error)
}

func TestFinishEventRunbookWaiterScopedToInstall(t *testing.T) {
	db := setupWaiterDB(t)
	a := &Activities{db: db}
	insertActiveWaiter(t, db, "waiter-1", "install-a", "step-1")

	_, err := a.FinishEventRunbookWaiter(context.Background(), FinishEventRunbookWaiterRequest{
		WorkflowStepID: "step-1", InstallID: "install-b", Status: app.EventRunbookWaiterStatusCancelled,
	})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	var status string
	require.NoError(t, db.Raw(`SELECT status FROM event_runbook_waiters WHERE id = 'waiter-1'`).Scan(&status).Error)
	require.Equal(t, "active", status)
}

func TestFinishEventRunbookWaiterCancelsOwnWaiter(t *testing.T) {
	db := setupWaiterDB(t)
	a := &Activities{db: db}
	insertActiveWaiter(t, db, "waiter-1", "install-a", "step-1")

	status, err := a.FinishEventRunbookWaiter(context.Background(), FinishEventRunbookWaiterRequest{
		WorkflowStepID: "step-1", InstallID: "install-a", Status: app.EventRunbookWaiterStatusCancelled,
	})
	require.NoError(t, err)
	require.Equal(t, app.EventRunbookWaiterStatusCancelled, status)
}

func TestFinishEventRunbookWaiterRejectsNonTerminalStatus(t *testing.T) {
	db := setupWaiterDB(t)
	a := &Activities{db: db}

	_, err := a.FinishEventRunbookWaiter(context.Background(), FinishEventRunbookWaiterRequest{
		WorkflowStepID: "step-1", InstallID: "install-a", Status: app.EventRunbookWaiterStatusMatched,
	})
	require.ErrorContains(t, err, "invalid terminal waiter status")
}
