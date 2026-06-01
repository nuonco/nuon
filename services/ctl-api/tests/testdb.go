package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/services/config"
	"github.com/nuonco/nuon/pkg/workflows/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	psqlmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// DBConfig holds just the database connection fields we need.
// Uses the same config tags as internal.Config so it picks up the registered defaults.
type DBConfig struct {
	DBHost     string `config:"db_host"`
	DBPort     string `config:"db_port"`
	DBUser     string `config:"db_user"`
	DBPassword string `config:"db_password"`
	DBSSLMode  string `config:"db_ssl_mode"`
	DBName     string `config:"db_name"`
}

// SkipIfNotIntegration skips the test if INTEGRATION != "true".
// Call this at the start of TestXxxSuite functions.
func SkipIfNotIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
	}
}

// LoadDBConfig loads the database config from environment variables.
func LoadDBConfig() (DBConfig, error) {
	var cfg DBConfig
	if err := config.LoadInto(nil, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to load db config: %w", err)
	}
	if cfg.DBName == "" {
		return cfg, fmt.Errorf("DB_NAME must be set in the environment")
	}
	return cfg, nil
}

// CreateAndMigrateDatabase drops and recreates the test database, then runs migrations.
// Called by the testsetup binary before tests run.
func CreateAndMigrateDatabase(cfg DBConfig) error {
	// Connect to the default 'postgres' database to create test database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	defer sqlDB.Close()

	// Terminate existing connections and drop the database
	db.Exec(fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s'", cfg.DBName))
	db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", cfg.DBName))

	// Create fresh database
	if err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName)).Error; err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}

	// Run migrations
	if err := MigrateTestDatabase(cfg); err != nil {
		return fmt.Errorf("failed to migrate test database: %w", err)
	}

	return nil
}

// MigrateTestDatabase connects to the test database and runs all migrations.
func MigrateTestDatabase(cfg DBConfig) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	defer sqlDB.Close()

	// Enable required extensions
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS hstore").Error; err != nil {
		return fmt.Errorf("failed to create hstore extension: %w", err)
	}

	if err := runMigrator(context.Background(), db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// BaseDBTestSuite provides the base test suite for database-backed tests.
// Embed this in your test suites. The database must already exist and be migrated
// (via the testsetup binary) before tests run.
//
// Tests rely on unique names for data isolation — no truncation is needed.
//
// Example:
//
//	type MyTestSuite struct {
//	    tests.BaseDBTestSuite
//	    // your fields
//	}
//
//	func TestMySuite(t *testing.T) {
//	    tests.SkipIfNotIntegration(t)
//	    suite.Run(t, new(MyTestSuite))
//	}
//
//	func (s *MyTestSuite) SetupSuite() {
//	    s.BaseDBTestSuite.SetupSuite()
//	    // create your fx app and get DB
//	    s.SetDB(db)
//	    s.SetCHDB(chDB) // optional
//	}
type BaseDBTestSuite struct {
	suite.Suite
	db   *gorm.DB
	chDB *gorm.DB
}

// SetupSuite is a no-op. The database is created by the testsetup binary before tests run.
func (s *BaseDBTestSuite) SetupSuite() {}

// SetDB stores the PostgreSQL database connection.
func (s *BaseDBTestSuite) SetDB(db *gorm.DB) {
	s.db = db
}

// DB returns the PostgreSQL database connection.
func (s *BaseDBTestSuite) DB() *gorm.DB {
	return s.db
}

// SetCHDB stores the ClickHouse database connection.
func (s *BaseDBTestSuite) SetCHDB(db *gorm.DB) {
	s.chDB = db
}

// CHDB returns the ClickHouse database connection.
func (s *BaseDBTestSuite) CHDB() *gorm.DB {
	return s.chDB
}

// SetupTest is a no-op. Tests use unique names for data isolation.
func (s *BaseDBTestSuite) SetupTest() {}

func runMigrator(ctx context.Context, db *gorm.DB) error {
	testConfig := &internal.Config{
		Config: worker.Config{
			Env:                             config.Development,
			ServiceName:                     "ctl-api-test",
			GitRef:                          "test",
			Version:                         "test",
			LogLevel:                        "error",
			TemporalHost:                    "localhost:7233",
			TemporalTaskQueue:               "test",
			TemporalMaxConcurrentActivities: 1,
			HostIP:                          "localhost",
		},
		ServiceType: "test",
	}

	logger := zap.NewNop()
	v := validator.New()
	metricsWriter, err := metrics.New(
		v,
		metrics.WithDisable(true),
		metrics.WithLogger(logger),
	)
	if err != nil {
		return fmt.Errorf("failed to create metrics writer: %w", err)
	}

	models := psql.AllModels()
	acctClient := account.New(account.Params{
		Cfg:             testConfig,
		AnalyticsClient: nil, // Not needed for test migrations
		DB:              db,
		V:               v,
		AuthzClient:     nil, // Not needed for test migrations
	})

	psqlMigs := psqlmigrations.New(psqlmigrations.Params{
		AcctClient: acctClient,
	})

	migrator := migrations.New(migrations.Params{
		Models:       models,
		Migrations:   psqlMigs.All(),
		MigrationsDB: db,
		DB:           db,
		DBType:       "postgres",
		L:            logger,
		Cfg:          testConfig,
		MW:           metricsWriter,
		Opts:         migrations.NewOpts(),
		TableOpts:    map[string]string{},
	})

	// Execute migrations
	if err := migrator.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	return nil
}
