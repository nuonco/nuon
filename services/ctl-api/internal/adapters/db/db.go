package db

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

// database represents the set of configuration options for creating a database connection. If UseIAM is set, we will
// automatically create a database token using the AWS RDS api.
type database struct {
	User    string `validate:"required"`
	Host    string `validate:"required"`
	Name    string `validate:"required"`
	Port    string `validate:"required"`
	SSLMode string `validate:"required"`

	// required for IAM auth
	PasswordFn func(context.Context, database) (string, error)
	Region     string `validate:"required"`

	// required for local auth
	Password string

	Logger zapgorm2.Logger `validate:"required"`

	MetricsWriter metrics.Writer `validate:"required"`

	pool          *pgxpool.Pool
	poolCtx       context.Context
	poolCtxCancel context.CancelFunc
}

func (d *database) Validate(v *validator.Validate) error {
	if err := v.Struct(d); err != nil {
		return fmt.Errorf("unable to validate database: %w", err)
	}

	return nil
}

func (c *database) connCfg() (*pgx.ConnConfig, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.Host,
		c.User,
		c.Password,
		c.Name,
		c.Port,
		c.SSLMode)

	connCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	return connCfg, nil
}

func New(v *validator.Validate,
	l *zap.Logger,
	metricsWriter metrics.Writer,
	lc fx.Lifecycle,
	cfg *internal.Config) (*gorm.DB, error) {

	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	database := &database{
		Logger:        zapgorm2.New(l),
		PasswordFn:    FetchIamTokenPassword,
		Host:          cfg.DBHost,
		User:          cfg.DBUser,
		Name:          cfg.DBName,
		Port:          cfg.DBPort,
		SSLMode:       cfg.DBSSLMode,
		Region:        cfg.DBRegion,
		MetricsWriter: metricsWriter,
		poolCtx:       ctx,
		poolCtxCancel: cancelFn,
	}
	if cfg.DBPassword != "" {
		database.PasswordFn = nil
		database.Password = cfg.DBPassword
	}
	if err := database.Validate(v); err != nil {
		return nil, err
	}

	// create pool
	pool, err := database.createPool()
	if err != nil {
		return nil, fmt.Errorf("unable to create database pool: %w", err)
	}
	database.pool = pool

	// create postgres config and setup gorm
	postgresCfg := postgres.Config{
		Conn: stdlib.OpenDBFromPool(pool),
	}
	gormCfg := &gorm.Config{
		Logger:         database.Logger,
		TranslateError: true,
	}
	db, err := gorm.Open(postgres.New(postgresCfg), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// register plugins
	if err := database.registerPlugins(db); err != nil {
		return nil, fmt.Errorf("unable to register plugins: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			database.startPoolBackgroundJob()
			return nil
		},
		OnStop: func(_ context.Context) error {
			database.stopPoolBackgroundJob()
			return nil
		},
	})

	return db, err
}
