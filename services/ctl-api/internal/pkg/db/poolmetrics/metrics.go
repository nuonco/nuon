package poolmetrics

import (
	"context"
	"database/sql"

	"github.com/XSAM/otelsql"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
)

type Params struct {
	fx.In
	Metrics *Metrics `optional:"true"`
}

type Metrics struct {
	provider metric.MeterProvider
}

func New(provider metric.MeterProvider) *Metrics {
	return &Metrics{provider: provider}
}

func (m *Metrics) RegisterPostgres(role string, pool *pgxpool.Pool) error {
	return otelpgx.RecordStats(pool,
		otelpgx.WithStatsMeterProvider(m.provider),
		otelpgx.WithStatsAttributes(attribute.String("db.client.connection.pool.name", role)),
	)
}

func (m *Metrics) RegisterClickHouse(lc fx.Lifecycle, db *sql.DB) error {
	registration, err := otelsql.RegisterDBStatsMetrics(db,
		otelsql.WithMeterProvider(m.provider),
		otelsql.WithAttributes(
			attribute.String("db.system.name", "clickhouse"),
			attribute.String("db.client.connection.pool.name", "primary"),
		),
	)
	if err != nil {
		return err
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return registration.Unregister() }})
	return nil
}
