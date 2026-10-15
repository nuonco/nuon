package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func seedRunnableInstall(ctx context.Context, t *testing.T, db *gorm.DB, installID string) {
	t.Helper()
	group := &app.RunnerGroup{
		OwnerID:   installID,
		OwnerType: "installs",
		Type:      app.RunnerGroupTypeInstall,
		Platform:  app.AppRunnerTypeAWS,
	}
	require.NoError(t, db.WithContext(ctx).Create(group).Error)
	runner := &app.Runner{
		RunnerGroupID:     group.ID,
		Name:              "runner-" + installID,
		DisplayName:       "runner",
		Status:            app.RunnerStatusActive,
		StatusDescription: "active",
	}
	require.NoError(t, db.WithContext(ctx).Create(runner).Error)
	require.NoError(t, db.WithContext(ctx).Create(&app.RunnerProcess{
		RunnerID:        runner.ID,
		Type:            app.RunnerProcessTypeInstall,
		CompositeStatus: app.CompositeStatus{Status: app.Status(app.RunnerProcessStatusActive)},
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&app.Queue{
		OwnerID:     installID,
		OwnerType:   "installs",
		Name:        helpers.InstallActionWorkflowsQueueName,
		MaxDepth:    100,
		MaxInFlight: 1,
	}).Error)
}

func assertCompositeError(t *testing.T, expected, actual *compositeerrors.CompositeErrorData) {
	t.Helper()
	require.NotNil(t, actual)
	assert.Equal(t, expected.Version, actual.Version)
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Severity, actual.Severity)
	assert.Equal(t, expected.Message, actual.Message)
}
