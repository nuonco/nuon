package hooks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func TestRetryDBRead(t *testing.T) {
	t.Run("succeeds first attempt without retrying", func(t *testing.T) {
		calls := 0
		err := retryDBRead(context.Background(), func() error {
			calls++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 1, calls)
	})

	t.Run("recovers from transient failure", func(t *testing.T) {
		calls := 0
		err := retryDBRead(context.Background(), func() error {
			calls++
			if calls < 2 {
				return errors.New("transient connection failure")
			}
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 2, calls)
	})

	t.Run("gives up after bounded attempts", func(t *testing.T) {
		calls := 0
		persistent := errors.New("db down")
		err := retryDBRead(context.Background(), func() error {
			calls++
			return persistent
		})
		require.ErrorIs(t, err, persistent)
		assert.Equal(t, dbReadRetryAttempts, calls)
	})

	t.Run("record not found is a domain outcome, never retried", func(t *testing.T) {
		calls := 0
		err := retryDBRead(context.Background(), func() error {
			calls++
			return gorm.ErrRecordNotFound
		})
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.Equal(t, 1, calls)
	})

	t.Run("cancelled context stops retrying", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		err := retryDBRead(ctx, func() error {
			calls++
			cancel()
			return errors.New("transient")
		})
		require.Error(t, err)
		assert.Equal(t, 1, calls)
	})
}

// Proves the workflow status read behind resolveFlowCompletionOutcome
// survives a transient DB failure instead of dropping the notification, and
// that exhausted retries still surface an error (callers must not publish a
// false "succeeded").
func TestResolveFlowCompletionOutcomeRetriesTransientReads(t *testing.T) {
	newDB := func(t *testing.T, status app.CompositeStatus) *gorm.DB {
		t.Helper()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&workflowStatusRow{}))
		require.NoError(t, db.Create(&workflowStatusRow{
			ID:     "wfl_1",
			Status: status,
		}).Error)
		return db
	}

	event := signal.SignalPhaseEvent{
		SignalType: signalTypeExecuteWorkflow,
		Phase:      signal.SignalPhaseExecute,
		WorkflowID: "wfl_1",
	}

	t.Run("transient failure recovers and rewrites domain outcome", func(t *testing.T) {
		db := newDB(t, app.CompositeStatus{
			Status:                 app.StatusCancelled,
			StatusHumanDescription: "cancelled by user",
		})
		failures := 1
		require.NoError(t, db.Callback().Query().Before("gorm:query").
			Register("fail_transiently", func(tx *gorm.DB) {
				if failures > 0 {
					failures--
					_ = tx.AddError(errors.New("transient connection failure"))
				}
			}))

		outcome := signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}
		suppress, err := resolveFlowCompletionOutcome(context.Background(), db, event, &outcome)
		require.NoError(t, err)
		assert.False(t, suppress)
		assert.Equal(t, signal.SignalStatusCancelled, outcome.Status)
		assert.Equal(t, "cancelled by user", outcome.ErrMessage)
	})

	t.Run("exhausted retries still block publishing", func(t *testing.T) {
		db := newDB(t, app.CompositeStatus{Status: app.StatusSuccess})
		require.NoError(t, db.Callback().Query().Before("gorm:query").
			Register("fail_always", func(tx *gorm.DB) {
				_ = tx.AddError(errors.New("db down"))
			}))

		outcome := signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}
		_, err := resolveFlowCompletionOutcome(context.Background(), db, event, &outcome)
		require.Error(t, err)
	})
}
