package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nuonco/nuon/pkg/nuonauth"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// emptySchemaDB connects to the lane database with search_path set to a new, empty schema,
// so o_auth_clients is missing without touching the shared tables.
func emptySchemaDB(t *testing.T) *gorm.DB {
	cfg, err := tests.LoadDBConfig()
	require.NoError(t, err)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	schema := "auth_startup_" + strings.ToLower(domains.NewOAuthClientID())
	require.NoError(t, admin.Exec(fmt.Sprintf("CREATE SCHEMA %s", schema)).Error)

	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		admin.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE", schema))
		if sqlDB, err := admin.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return db
}

func TestStartCLIClient(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	prevInterval := cliClientRetryInterval
	cliClientRetryInterval = 100 * time.Millisecond
	t.Cleanup(func() { cliClientRetryInterval = prevInterval })

	newService := func(t *testing.T, db *gorm.DB) *service {
		s := &service{db: db, l: zap.NewNop(), stopCleanup: make(chan struct{})}
		t.Cleanup(func() { close(s.stopCleanup) })
		return s
	}

	t.Run("retries until o_auth_clients exists", func(t *testing.T) {
		db := emptySchemaDB(t)
		s := newService(t, db)

		require.NoError(t, s.startCLIClient(context.Background()))

		require.NoError(t, db.AutoMigrate(&app.OAuthClient{}))
		require.Eventually(t, func() bool {
			var client app.OAuthClient
			return db.Where(&app.OAuthClient{ID: nuonauth.CLIClientID}).First(&client).Error == nil
		}, 10*time.Second, 100*time.Millisecond)
	})

	t.Run("fails startup on other errors", func(t *testing.T) {
		s := newService(t, emptySchemaDB(t))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		require.Error(t, s.startCLIClient(ctx))
	})
}
