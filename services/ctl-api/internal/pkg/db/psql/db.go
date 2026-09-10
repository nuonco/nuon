package psql

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/querycollector"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
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

	// pool configuration
	MaxConnections int32
	Role           string

	Logger zapgorm2.Logger `validate:"required"`

	MetricsWriter  metrics.Writer `validate:"required"`
	QueryCollector *querycollector.Collector

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
	var dsn string
	if c.PasswordFn != nil {
		dsn = fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s",
			c.Host,
			c.User,
			c.Name,
			c.Port,
			c.SSLMode)
	} else {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			c.Host,
			c.User,
			c.Password,
			c.Name,
			c.Port,
			c.SSLMode)
	}

	connCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	return connCfg, nil
}

func New(v *validator.Validate,
	l zapgorm2.Logger,
	metricsWriter metrics.Writer,
	lc fx.Lifecycle,
	cfg *internal.Config,
	qc *querycollector.Collector,
) (*gorm.DB, error) {
	primary, err := newDatabase(cfg, l, metricsWriter, qc, cfg.DBHost)
	if err != nil {
		return nil, fmt.Errorf("unable to build primary database config: %w", err)
	}
	primary.Role = "primary"
	if err := primary.Validate(v); err != nil {
		return nil, err
	}

	primaryPool, err := primary.createPool()
	if err != nil {
		return nil, fmt.Errorf("unable to create primary pool: %w", err)
	}
	primary.pool = primaryPool
	primarySQL := stdlib.OpenDBFromPool(primaryPool)

	var replica *database
	connPool := routing.NewConnPool(primarySQL, nil)

	if cfg.DBReplicaEnabled && cfg.DBGormReplicaHost != "" {
		replica, err = newDatabase(cfg, l, metricsWriter, qc, cfg.DBGormReplicaHost)
		if err != nil {
			return nil, fmt.Errorf("unable to build replica database config: %w", err)
		}
		replica.Role = "replica"
		if err := replica.Validate(v); err != nil {
			return nil, fmt.Errorf("unable to validate replica database: %w", err)
		}

		replicaPool, err := replica.createPool()
		if err != nil {
			return nil, fmt.Errorf("unable to create replica pool: %w", err)
		}
		replica.pool = replicaPool

		connPool = routing.NewConnPool(primarySQL, stdlib.OpenDBFromPool(replicaPool))
		l.ZapLogger.Info("replica pool initialized",
			zap.String("host", cfg.DBGormReplicaHost),
			zap.Bool("bypass_opt_in", cfg.DBReplicaBypassOptIn),
		)
	} else {
		l.ZapLogger.Info("replica routing disabled, primary-only pool",
			zap.Bool("enabled", cfg.DBReplicaEnabled),
			zap.String("host", cfg.DBGormReplicaHost),
		)
	}

	postgresCfg := postgres.Config{Conn: primarySQL}
	gormCfg := primary.gormConfig()

	db, err := gorm.Open(postgres.New(postgresCfg), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	db.ConnPool = connPool

	if err := primary.registerPlugins(db); err != nil {
		return nil, fmt.Errorf("unable to register plugins: %w", err)
	}

	if err := db.Use(&routing.Plugin{
		ACL:         replicaACL(db),
		BypassOptIn: cfg.DBReplicaBypassOptIn,
		Logger:      l.ZapLogger,
	}); err != nil {
		return nil, fmt.Errorf("unable to register routing plugin: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			primary.startPoolBackgroundJob()
			if replica != nil {
				replica.startPoolBackgroundJob()
			}
			return nil
		},
		OnStop: func(_ context.Context) error {
			primary.stopPoolBackgroundJob()
			if replica != nil {
				replica.stopPoolBackgroundJob()
			}
			return nil
		},
	})

	return db, nil
}

// NewReplica errors if DBReplicaHost is empty so callers can't silently
// fall back to the primary.
func NewReplica(v *validator.Validate,
	l zapgorm2.Logger,
	metricsWriter metrics.Writer,
	lc fx.Lifecycle,
	cfg *internal.Config,
	qc *querycollector.Collector,
) (*gorm.DB, error) {
	if cfg.DBReplicaHost == "" {
		return nil, fmt.Errorf("db_replica_host must be set to use the read replica")
	}
	return open(v, l, metricsWriter, lc, cfg, qc, cfg.DBReplicaHost)
}

func newDatabase(cfg *internal.Config, l zapgorm2.Logger, metricsWriter metrics.Writer, qc *querycollector.Collector, host string) (*database, error) {
	ctx, cancelFn := context.WithCancel(context.Background())

	d := &database{
		Logger:         l,
		Host:           host,
		User:           cfg.DBUser,
		Name:           cfg.DBName,
		Port:           cfg.DBPort,
		SSLMode:        cfg.DBSSLMode,
		Region:         cfg.DBRegion,
		MaxConnections: cfg.DBMaxConnections,
		MetricsWriter:  metricsWriter,
		QueryCollector: qc,
		poolCtx:        ctx,
		poolCtxCancel:  cancelFn,
	}

	switch {
	case cfg.DBPassword != "":
		d.Password = cfg.DBPassword
	case cfg.DBUseIAM && cfg.IsGCP():
		d.PasswordFn = FetchGcpCloudSqlPassword
	case cfg.DBUseIAM:
		d.PasswordFn = FetchIamTokenPassword
	}

	return d, nil
}

func open(v *validator.Validate,
	l zapgorm2.Logger,
	metricsWriter metrics.Writer,
	lc fx.Lifecycle,
	cfg *internal.Config,
	qc *querycollector.Collector,
	host string,
) (*gorm.DB, error) {
	d, err := newDatabase(cfg, l, metricsWriter, qc, host)
	if err != nil {
		return nil, err
	}
	if err := d.Validate(v); err != nil {
		return nil, err
	}

	pool, err := d.createPool()
	if err != nil {
		return nil, fmt.Errorf("unable to create database pool: %w", err)
	}
	d.pool = pool

	postgresCfg := d.postgresConfig(pool)
	gormCfg := d.gormConfig()

	db, err := gorm.Open(postgres.New(postgresCfg), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := d.registerPlugins(db); err != nil {
		return nil, fmt.Errorf("unable to register plugins: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			d.startPoolBackgroundJob()
			return nil
		},
		OnStop: func(_ context.Context) error {
			d.stopPoolBackgroundJob()
			return nil
		},
	})

	return db, err
}
