package enqueuer

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func TestEnqueueInlineQuarantinesSignalWithMissingQueue(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE queue_signals (
			id TEXT PRIMARY KEY,
			updated_at DATETIME,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			queue_id TEXT NOT NULL,
			owner_id TEXT NOT NULL,
			owner_type TEXT NOT NULL,
			status JSON NOT NULL,
			type TEXT NOT NULL,
			enqueued BOOLEAN NOT NULL DEFAULT FALSE
		)
	`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE queues (id TEXT PRIMARY KEY, deleted_at INTEGER NOT NULL DEFAULT 0)`).Error)

	mw, err := metrics.New(validator.New(), metrics.WithDisable(true), metrics.WithLogger(zap.NewNop()))
	require.NoError(t, err)

	qs := &app.QueueSignal{
		ID:        "qsg00000000000000000000000",
		QueueID:   "que00000000000000000000000",
		OwnerID:   "ins00000000000000000000000",
		OwnerType: "installs",
		Status:    app.NewCompositeStatus(context.Background(), app.StatusQueued),
		Type:      signal.SignalType("test-signal"),
	}
	require.NoError(t, db.Exec(
		"INSERT INTO queue_signals (id, queue_id, owner_id, owner_type, status, type) VALUES (?, ?, ?, ?, ?, ?)",
		qs.ID, qs.QueueID, qs.OwnerID, qs.OwnerType, &qs.Status, qs.Type,
	).Error)

	e := &Enqueuer{db: db, l: zap.NewNop(), mw: mw}
	err = e.EnqueueInline(context.Background(), qs.ID, EnqueueSourceSweep)
	require.ErrorIs(t, err, errOrphanedQueueSignal)

	var got app.QueueSignal
	require.NoError(t, db.Unscoped().First(&got, "id = ?", qs.ID).Error)
	require.NotZero(t, got.DeletedAt)
	require.False(t, got.Enqueued)
	require.Equal(t, app.StatusError, got.Status.Status)
	require.Equal(t, "parent queue not found; signal quarantined before enqueue", got.Status.StatusHumanDescription)
}
