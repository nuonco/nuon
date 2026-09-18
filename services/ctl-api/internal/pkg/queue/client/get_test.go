package client

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestDefaultQueueLookupIncludesEmptyName(t *testing.T) {
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
	var queue app.Queue
	result := client.defaultQueueByOwnerQuery(context.Background(), "owner-id", "app_branches", "").First(&queue)
	require.NoError(t, result.Error)
	sql := strings.ToLower(result.Statement.SQL.String())
	require.Contains(t, sql, `"name"`)
	require.Contains(t, result.Statement.Vars, "")
}
