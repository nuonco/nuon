package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestReconcileQueueCapacity(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=unused",
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
		DryRun:                 true,
	})
	require.NoError(t, err)

	client := &Client{db: db}
	existing := &app.Queue{ID: "que00000000000000000000000", MaxInFlight: 2, MaxDepth: 10}
	req := &CreateQueueRequest{MaxInFlight: 5, MaxDepth: 50}

	require.NoError(t, client.reconcileQueueCapacity(context.Background(), existing, req))
	require.Equal(t, 5, existing.MaxInFlight)
	require.Equal(t, 50, existing.MaxDepth)
}
