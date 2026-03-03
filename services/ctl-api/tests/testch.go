package tests

import (
	"context"
	"fmt"
	"time"

	clickhousecore "github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	chpkg "github.com/nuonco/nuon/pkg/gorm/clickhouse"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/services/config"
	"github.com/nuonco/nuon/pkg/workflows/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch"
	chmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// CHConfig holds the ClickHouse connection fields.
// Uses the same config tags as internal.Config so it picks up the registered defaults.
type CHConfig struct {
	Host     string `config:"clickhouse_db_host"`
	Port     string `config:"clickhouse_db_port"`
	User     string `config:"clickhouse_db_user"`
	Password string `config:"clickhouse_db_password"`
	Name     string `config:"clickhouse_db_name"`
	UseTLS   bool   `config:"clickhouse_db_use_tls"`

	ReadTimeout  time.Duration `config:"clickhouse_db_read_timeout"`
	WriteTimeout time.Duration `config:"clickhouse_db_write_timeout"`
	DialTimeout  time.Duration `config:"clickhouse_db_dial_timeout"`
}

// LoadCHConfig loads the ClickHouse config from environment variables.
func LoadCHConfig() (CHConfig, error) {
	var cfg CHConfig
	if err := config.LoadInto(nil, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to load clickhouse config: %w", err)
	}
	if cfg.Name == "" {
		return cfg, fmt.Errorf("CLICKHOUSE_DB_NAME must be set in the environment")
	}
	return cfg, nil
}

// CreateAndMigrateCHDatabase drops and recreates the ClickHouse test database, then runs migrations.
// Called by the testsetup binary before tests run.
func CreateAndMigrateCHDatabase(chCfg CHConfig) error {
	// Connect to the default database to create our test database
	defaultOpts := &clickhousecore.Options{
		Addr: []string{fmt.Sprintf("%s:%s", chCfg.Host, chCfg.Port)},
		Auth: clickhousecore.Auth{
			Database: "default",
			Username: chCfg.User,
			Password: chCfg.Password,
		},
		Settings: clickhousecore.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: chCfg.DialTimeout,
		ReadTimeout: chCfg.ReadTimeout,
		Compression: &clickhousecore.Compression{
			Method: clickhousecore.CompressionLZ4,
		},
	}

	defaultPool := clickhousecore.OpenDB(defaultOpts)
	defaultDB, err := gorm.Open(chpkg.New(chpkg.Config{Conn: defaultPool}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to clickhouse default database: %w", err)
	}

	sqlDB, err := defaultDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get clickhouse sql.DB: %w", err)
	}
	defer sqlDB.Close()

	// Drop and recreate database
	defaultDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s ON CLUSTER simple", chCfg.Name))

	if err := defaultDB.Exec(fmt.Sprintf("CREATE DATABASE %s ON CLUSTER simple", chCfg.Name)).Error; err != nil {
		return fmt.Errorf("failed to create clickhouse test database: %w", err)
	}

	// Run migrations
	if err := MigrateTestCHDatabase(chCfg); err != nil {
		return fmt.Errorf("failed to migrate clickhouse test database: %w", err)
	}

	return nil
}

// MigrateTestCHDatabase connects to the ClickHouse test database and runs migrations.
// CH migration state is tracked in PostgreSQL, so we need both connections.
func MigrateTestCHDatabase(chCfg CHConfig) error {
	// Connect to ClickHouse target database
	chOpts := &clickhousecore.Options{
		Addr: []string{fmt.Sprintf("%s:%s", chCfg.Host, chCfg.Port)},
		Auth: clickhousecore.Auth{
			Database: chCfg.Name,
			Username: chCfg.User,
			Password: chCfg.Password,
		},
		Settings: clickhousecore.Settings{
			"max_execution_time":               60,
			"async_insert":                     1,
			"wait_for_async_insert":            1,
			"async_insert_busy_timeout_min_ms": 200,
			"async_insert_busy_timeout_max_ms": 1000,
			"distributed_ddl_task_timeout":     600,
		},
		DialTimeout: chCfg.DialTimeout,
		ReadTimeout: chCfg.ReadTimeout,
		Compression: &clickhousecore.Compression{
			Method: clickhousecore.CompressionLZ4,
		},
	}

	chPool := clickhousecore.OpenDB(chOpts)
	chDB, err := gorm.Open(chpkg.New(chpkg.Config{Conn: chPool}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to clickhouse test database: %w", err)
	}

	chSqlDB, err := chDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get clickhouse sql.DB: %w", err)
	}
	defer chSqlDB.Close()

	// Connect to PostgreSQL for migration tracking
	var psqlCfg DBConfig
	if err := config.LoadInto(nil, &psqlCfg); err != nil {
		return fmt.Errorf("failed to load psql config for CH migration tracking: %w", err)
	}

	psqlDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		psqlCfg.DBHost, psqlCfg.DBPort, psqlCfg.DBUser, psqlCfg.DBPassword, psqlCfg.DBName, psqlCfg.DBSSLMode)

	psqlDB, err := gorm.Open(postgres.Open(psqlDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres for CH migration tracking: %w", err)
	}

	psqlSqlDB, err := psqlDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get postgres sql.DB: %w", err)
	}
	defer psqlSqlDB.Close()

	return runCHMigrator(context.Background(), chDB, psqlDB)
}

func runCHMigrator(ctx context.Context, chDB, psqlDB *gorm.DB) error {
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

	l := zap.NewNop()
	v := validator.New()
	metricsWriter, err := metrics.New(
		v,
		metrics.WithDisable(true),
		metrics.WithLogger(l),
	)
	if err != nil {
		return fmt.Errorf("failed to create metrics writer: %w", err)
	}

	chMigs := chmigrations.New(chmigrations.Params{})

	opts := migrations.NewOpts()
	opts.CreateViewSQLTmpl = "CREATE OR REPLACE VIEW %s ON CLUSTER simple AS %s"

	migrator := migrations.New(migrations.Params{
		Models:       ch.AllModels(),
		Migrations:   chMigs.All(),
		MigrationsDB: psqlDB,
		DB:           chDB,
		DBType:       "ch",
		L:            l,
		Cfg:          testConfig,
		MW:           metricsWriter,
		Opts:         opts,
		TableOpts: map[string]string{
			"gorm:table_cluster_options": "on cluster simple",
		},
	})

	if err := migrator.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute clickhouse migrations: %w", err)
	}

	return nil
}
