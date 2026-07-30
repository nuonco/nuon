package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestHoldCompletionCallbacks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&workflowStatusRow{}))

	activities := &Activities{db: db}
	ctx := context.Background()

	require.NoError(t, db.Create(&workflowStatusRow{
		ID: "wfl_1",
		Status: app.CompositeStatus{
			Status: app.StatusFailedPendingRetry,
		},
	}).Error)

	hold, err := activities.holdCompletionCallbacks(ctx, "wfl_1")
	require.NoError(t, err)
	assert.True(t, hold)

	require.NoError(t, db.Save(&workflowStatusRow{
		ID: "wfl_1",
		Status: app.CompositeStatus{
			Status: app.StatusSuccess,
		},
	}).Error)

	hold, err = activities.holdCompletionCallbacks(ctx, "wfl_1")
	require.NoError(t, err)
	assert.False(t, hold)
}
