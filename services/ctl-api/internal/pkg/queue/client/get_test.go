package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGetQueueByOwnerFiltersToDefaultQueue(t *testing.T) {
	db, err := gorm.Open(
		postgres.New(postgres.Config{DSN: "postgres://ctl_api@127.0.0.1:5432/ctl_api"}),
		&gorm.Config{DryRun: true, DisableAutomaticPing: true},
	)
	require.NoError(t, err)

	var sql string
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("test:capture", func(tx *gorm.DB) {
		sql = tx.Statement.SQL.String()
	}))

	_, _ = (&Client{db: db}).GetQueueByOwner(context.Background(), "abr1", "app_branches")

	require.Contains(t, sql, `"name" = $3`, "owners have multiple named queues, so the lookup must pin the default one")
	require.Contains(t, sql, `"queues"."owner_id" = $1`)
}
