package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

func seedInstallActionWorkflowsQueue(t *testing.T, db *gorm.DB, ctx context.Context, installID string) {
	require.NoError(t, db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&app.QueueSignal{}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&app.Queue{
		OwnerID:     installID,
		OwnerType:   "installs",
		Name:        installhelpers.InstallActionWorkflowsQueueName,
		MaxDepth:    50,
		MaxInFlight: 10,
	}).Error)
}
