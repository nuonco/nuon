package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GetQueueSignals returns all queue signals from the DB, ordered by creation time.
func GetQueueSignals(t testing.TB, db *gorm.DB) []app.QueueSignal {
	var signals []app.QueueSignal
	res := db.Order("created_at ASC").Find(&signals)
	require.NoError(t, res.Error)
	return signals
}

// GetQueueSignalsByOwner returns queue signals for a specific owner.
func GetQueueSignalsByOwner(t testing.TB, db *gorm.DB, ownerID string) []app.QueueSignal {
	var signals []app.QueueSignal
	res := db.Where("owner_id = ?", ownerID).Order("created_at ASC").Find(&signals)
	require.NoError(t, res.Error)
	return signals
}

// ClearQueueSignals deletes all queue signals from the DB.
func ClearQueueSignals(t testing.TB, db *gorm.DB) {
	db.Unscoped().Where("1 = 1").Delete(&app.QueueSignal{})
}
