package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

type DBConfig struct {
	DBHost     string `config:"db_host"`
	DBPort     string `config:"db_port"`
	DBUser     string `config:"db_user"`
	DBPassword string `config:"db_password"`
	DBSSLMode  string `config:"db_ssl_mode"`
	DBName     string `config:"db_name"`
}

func SkipIfNotIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
	}
}

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

func ResetDatabase(cfg DBConfig) error {
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

	db.Exec(fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s'", cfg.DBName))
	db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", cfg.DBName))

	if err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName)).Error; err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}

	return nil
}

func CreateAndMigrateDatabase(cfg DBConfig) error {
	if err := ResetDatabase(cfg); err != nil {
		return err
	}

	if err := MigrateTestDatabase(cfg); err != nil {
		return fmt.Errorf("failed to migrate test database: %w", err)
	}

	return nil
}

func SchemaSnapshotDir() string {
	return os.Getenv("NUONTEST_PG_SCHEMA_DIR")
}

func SchemaSnapshotPath(cfg DBConfig) (string, bool) {
	dir := SchemaSnapshotDir()
	if dir == "" {
		return "", false
	}
	path := filepath.Join(dir, cfg.DBName+".sql")
	if _, err := os.Stat(path); err != nil {
		return "", false
	}
	return path, true
}

func RestoreDatabase(cfg DBConfig, snapshotPath string) error {
	if err := ResetDatabase(cfg); err != nil {
		return err
	}

	dump, err := os.Open(snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to open schema snapshot: %w", err)
	}
	defer dump.Close()

	restore := exec.Command("docker", "exec", "-i", pgDumpContainer(), "psql",
		"-U", cfg.DBUser, "-v", "ON_ERROR_STOP=1", "-d", cfg.DBName)
	restore.Stdin = dump
	restore.Stderr = os.Stderr
	if err := restore.Run(); err != nil {
		return fmt.Errorf("failed to restore schema snapshot: %w", err)
	}

	return nil
}

// why: DumpSchema writes a pg_dump of the migrated database into dir. The CI lane
// snapshots the freshly migrated schema so later runs restore it instead of
// replaying migrations; written atomically so a concurrent cache save never
// sees a partial file.
func DumpSchema(cfg DBConfig, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create schema snapshot dir: %w", err)
	}

	dump, err := os.CreateTemp(dir, cfg.DBName+".sql.*")
	if err != nil {
		return fmt.Errorf("failed to create temp schema snapshot: %w", err)
	}
	tmpPath := dump.Name()

	dumpCmd := exec.Command("docker", "exec", pgDumpContainer(), "pg_dump",
		"-U", cfg.DBUser, "--no-owner", "--no-privileges", cfg.DBName)
	dumpCmd.Stdout = dump
	dumpCmd.Stderr = os.Stderr
	if err := dumpCmd.Run(); err != nil {
		dump.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to dump schema snapshot: %w", err)
	}
	if err := dump.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close schema snapshot: %w", err)
	}

	if err := os.Rename(tmpPath, filepath.Join(dir, cfg.DBName+".sql")); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to finalize schema snapshot: %w", err)
	}

	return nil
}

func pgDumpContainer() string {
	if c := os.Getenv("NUONTEST_PG_CONTAINER"); c != "" {
		return c
	}
	return "postgres"
}

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

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS hstore").Error; err != nil {
		return fmt.Errorf("failed to create hstore extension: %w", err)
	}

	if err := runMigrator(context.Background(), db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

type BaseDBTestSuite struct {
	suite.Suite
	db   *gorm.DB
	chDB *gorm.DB
}

func (s *BaseDBTestSuite) SetupSuite() {}

func (s *BaseDBTestSuite) SetDB(db *gorm.DB) {
	s.db = db
}

func (s *BaseDBTestSuite) DB() *gorm.DB {
	return s.db
}

func (s *BaseDBTestSuite) SetCHDB(db *gorm.DB) {
	s.chDB = db
}

func (s *BaseDBTestSuite) CHDB() *gorm.DB {
	return s.chDB
}

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
		AnalyticsClient: nil,
		DB:              db,
		V:               v,
		AuthzClient:     nil,
	})

	psqlMigs := psqlmigrations.New(psqlmigrations.Params{
		AcctClient: acctClient,
		L:          logger,
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

	if err := migrator.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	return nil
}
