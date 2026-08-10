package ch

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"moul.io/zapgorm2"

	"github.com/nuonco/nuon/pkg/gorm/clickhouse"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/querycollector"
)

type Params struct {
	fx.In

	V              *validator.Validate
	L              zapgorm2.Logger
	Cfg            *internal.Config
	MetricsWriter  metrics.Writer
	QueryCollector *querycollector.Collector
}

// database represents the set of configuration options for creating a database connection. If UseIAM is set, we will
// automatically create a database token using the AWS RDS api.
type database struct {
	User     string `validate:"required"`
	Password string `validate:"required"`
	Host     string `validate:"required"`
	Port     string `validate:"required"`
	Name     string `validate:"required"`
	UseTLS   bool

	ReadTimeout  time.Duration `validate:"required"`
	WriteTimeout time.Duration `validate:"required"`
	DialTimeout  time.Duration `validate:"required"`

	Logger        zapgorm2.Logger `validate:"required"`
	MetricsWriter metrics.Writer  `validate:"required"`

	QueryCollector *querycollector.Collector
	Debug          bool
}

func (d *database) Validate(v *validator.Validate) error {
	if err := v.Struct(d); err != nil {
		return fmt.Errorf("unable to validate database: %w", err)
	}

	return nil
}

func New(params Params, lc fx.Lifecycle) (*gorm.DB, error) {
	database := &database{
		Logger:         params.L,
		Host:           params.Cfg.ClickhouseDBHost,
		Name:           params.Cfg.ClickhouseDBName,
		User:           params.Cfg.ClickhouseDBUser,
		Password:       params.Cfg.ClickhouseDBPassword,
		Port:           params.Cfg.ClickhouseDBPort,
		MetricsWriter:  params.MetricsWriter,
		QueryCollector: params.QueryCollector,
		Debug:          params.Cfg.LogLevel == "DEBUG",
		ReadTimeout:    params.Cfg.ClickhouseDBReadTimeout,
		WriteTimeout:   params.Cfg.ClickhouseDBWriteTimeout,
		DialTimeout:    params.Cfg.ClickhouseDBDialTimeout,
	}
	if err := database.Validate(params.V); err != nil {
		return nil, err
	}

	gormCfg := database.gormConfig()
	chOpts := database.chOptions()
	chGormCfg := database.chGormConfig(chOpts)

	db, err := gorm.Open(clickhouse.New(chGormCfg), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("unable to access database connection pool: %w", err)
	}

	// register plugins
	if err := database.registerPlugins(db); err != nil {
		return nil, fmt.Errorf("unable to register plugins: %w", err)
	}

	poolMetricsCtx, stopPoolMetrics := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			recordPoolMetrics(database.MetricsWriter, sqlDB.Stats())
			go func() {
				ticker := time.NewTicker(poolMetricsPeriod)
				defer ticker.Stop()
				for {
					select {
					case <-poolMetricsCtx.Done():
						return
					case <-ticker.C:
						recordPoolMetrics(database.MetricsWriter, sqlDB.Stats())
					}
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			stopPoolMetrics()
			return nil
		},
	})

	return db, err
}
