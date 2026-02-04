package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
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

const TestDBName = "ctl_api_test"

var createDBOnce sync.Once

// dbConfig holds just the database connection fields we need.
// Uses the same config tags as internal.Config so it picks up the registered defaults.
type dbConfig struct {
	DBHost     string `config:"db_host"`
	DBPort     string `config:"db_port"`
	DBUser     string `config:"db_user"`
	DBPassword string `config:"db_password"`
	DBSSLMode  string `config:"db_ssl_mode"`
}

// SkipIfNotIntegration skips the test if INTEGRATION != "true".
// Call this at the start of TestXxxSuite functions.
func SkipIfNotIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
	}
}

// CreateTestDatabase creates the test database if it doesn't exist and runs migrations.
// Uses the same config system as the service to get connection parameters.
func CreateTestDatabase() error {
	var cfg dbConfig
	if err := config.LoadInto(nil, &cfg); err != nil {
		return fmt.Errorf("failed to load db config: %w", err)
	}

	// Connect to the default 'postgres' database to create our test database
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

	// Check if database exists
	var exists bool
	err = db.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", TestDBName).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		if err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", TestDBName)).Error; err != nil {
			return fmt.Errorf("failed to create test database: %w", err)
		}
	}

	// Run migrations on the test database
	if err := migrateTestDatabase(cfg); err != nil {
		return fmt.Errorf("failed to migrate test database: %w", err)
	}

	return nil
}

// migrateTestDatabase connects to the test database and runs GORM AutoMigrate on all models.
func migrateTestDatabase(cfg dbConfig) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, TestDBName, cfg.DBSSLMode)

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

// TruncateAllTables truncates all tables in the database except service accounts.
// Service accounts are preserved to avoid re-running data migrations.
func TruncateAllTables(ctx context.Context, db *gorm.DB) error {
	models := psql.AllModels()

	tableNames := make([]string, 0, len(models))
	for _, model := range models {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(model); err != nil {
			return fmt.Errorf("failed to parse model: %w", err)
		}

		// Skip accounts table - we'll handle it separately
		if stmt.Schema.Table == "accounts" {
			continue
		}

		tableNames = append(tableNames, fmt.Sprintf(`"%s"`, stmt.Schema.Table))
	}

	// Truncate all tables except accounts
	sql := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE",
		strings.Join(tableNames, ", "))

	if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	// Delete non-service accounts from accounts table
	if err := db.WithContext(ctx).Exec("DELETE FROM accounts WHERE account_type <> 'service'").Error; err != nil {
		return fmt.Errorf("failed to clean accounts table: %w", err)
	}

	return nil
}

// BaseDBTestSuite provides automatic test database setup and truncation.
// Embed this in your test suites and call SetDB() in SetupSuite after creating your DB connection.
//
// Example:
//
//	type MyTestSuite struct {
//	    testdb.BaseDBTestSuite
//	    // your fields
//	}
//
//	func TestMySuite(t *testing.T) {
//	    testdb.SkipIfNotIntegration(t)
//	    suite.Run(t, new(MyTestSuite))
//	}
//
//	func (s *MyTestSuite) SetupSuite() {
//	    s.BaseDBTestSuite.SetupSuite() // creates test DB and sets env
//	    // create your fx app and get DB
//	    s.SetDB(db)
//	}
//
// Tables are automatically truncated before each test via SetupTest.
type BaseDBTestSuite struct {
	suite.Suite
	db *gorm.DB
}

// SetupSuite creates the test database if needed and sets DB_NAME env var.
// Call this at the start of your SetupSuite if you override it.
func (s *BaseDBTestSuite) SetupSuite() {
	// Create test database if it doesn't exist (only once per test run)
	createDBOnce.Do(func() {
		if err := CreateTestDatabase(); err != nil {
			s.T().Fatalf("failed to create test database: %v", err)
		}
	})

	// Set DB_NAME so fx app connects to test database
	os.Setenv("DB_NAME", TestDBName)
}

// SetDB stores the database connection for use in truncation.
// Call this in your SetupSuite after creating the DB connection.
func (s *BaseDBTestSuite) SetDB(db *gorm.DB) {
	s.db = db
}

// DB returns the database connection.
func (s *BaseDBTestSuite) DB() *gorm.DB {
	return s.db
}

// SetupTest truncates all tables before each test and re-runs migrations.
// If you override SetupTest in your suite, call s.BaseDBTestSuite.SetupTest() first.
func (s *BaseDBTestSuite) SetupTest() {
	if s.db == nil {
		s.T().Fatal("DB not set - call SetDB() in SetupSuite")
	}

	// Truncate all tables
	err := TruncateAllTables(context.Background(), s.db)
	require.NoError(s.T(), err)
}

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
		EvClient:        nil, // Not needed for test migrations
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
