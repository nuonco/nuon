package activities

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func insertWaiter(t *testing.T, db *gorm.DB, id, orgID, queueSignalID string, notifiedAt *time.Time) {
	t.Helper()
	require.NoError(t, db.Exec(`
		INSERT INTO event_runbook_waiters (
			id, created_at, updated_at, org_id, app_id, install_id, workflow_id, workflow_step_id,
			queue_signal_id, trigger_id, status, activated_at, notified_at
		) VALUES (?, datetime('now'), datetime('now'), ?, 'app-1', 'inst-1', 'wf-1', ?,
			?, 'trg-1', 'matched', datetime('now'), ?)
	`, id, orgID, id+"-step", queueSignalID, notifiedAt).Error)
}

func TestNotifyEventRunbookWaiterRejectsCrossOrgLookup(t *testing.T) {
	db := setupWaiterDB(t)
	a := &Activities{db: db}
	insertWaiter(t, db, "waiter-1", "org-a", "qs-1", nil)

	err := a.NotifyEventRunbookWaiter(context.Background(), NotifyEventRunbookWaiterRequest{
		WaiterID: "waiter-1", OrgID: "org-b", QueueSignalID: "qs-1",
	})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestNotifyEventRunbookWaiterRejectsQueueSignalMismatch(t *testing.T) {
	db := setupWaiterDB(t)
	a := &Activities{db: db}
	insertWaiter(t, db, "waiter-1", "org-a", "qs-1", nil)

	err := a.NotifyEventRunbookWaiter(context.Background(), NotifyEventRunbookWaiterRequest{
		WaiterID: "waiter-1", OrgID: "org-a", QueueSignalID: "qs-other",
	})
	var appErr *temporal.ApplicationError
	require.True(t, errors.As(err, &appErr))
	require.Equal(t, "waiter_queue_signal_mismatch", appErr.Type())
	require.True(t, appErr.NonRetryable())
}

func TestNotifyEventRunbookWaiterAlreadyNotifiedIsNoop(t *testing.T) {
	db := setupWaiterDB(t)
	a := &Activities{db: db}
	notified := time.Now().UTC()
	insertWaiter(t, db, "waiter-1", "org-a", "qs-1", &notified)

	require.NoError(t, a.NotifyEventRunbookWaiter(context.Background(), NotifyEventRunbookWaiterRequest{
		WaiterID: "waiter-1", OrgID: "org-a", QueueSignalID: "qs-1",
	}))
}
