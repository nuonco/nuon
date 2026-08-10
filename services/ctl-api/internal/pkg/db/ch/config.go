package ch

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	clickhousecore "github.com/ClickHouse/clickhouse-go/v2"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/gorm/clickhouse"
	"github.com/nuonco/nuon/pkg/metrics"
)

const poolMetricsPeriod = 10 * time.Second

func (c *database) gormConfig() *gorm.Config {
	return &gorm.Config{
		TranslateError: true,
		Logger:         c.Logger,
	}
}

func (c *database) chOptions() *clickhousecore.Options {
	var tlsCfg *tls.Config
	if c.UseTLS {
		tlsCfg = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	opts := &clickhousecore.Options{
		Addr: []string{
			fmt.Sprintf("%s:%s", c.Host, c.Port),
		},
		Auth: clickhousecore.Auth{
			Database: c.Name,
			Username: c.User,
			Password: c.Password,
		},
		TLS: tlsCfg,
		Settings: clickhousecore.Settings{
			"max_execution_time":               5,
			"async_insert":                     1,
			"wait_for_async_insert":            1,
			"async_insert_busy_timeout_min_ms": 200,
			"async_insert_busy_timeout_max_ms": 1000,
			"distributed_ddl_task_timeout":     600,
		},
		DialTimeout: c.DialTimeout,
		ReadTimeout: c.ReadTimeout,
		Compression: &clickhousecore.Compression{
			Method: clickhousecore.CompressionLZ4,
		},
		Debug: c.Debug,
	}

	return opts
}

func (c *database) chGormConfig(opts *clickhousecore.Options) clickhouse.Config {
	pool := clickhousecore.OpenDB(opts)

	// database/sql defaults (MaxOpenConns=unlimited, MaxIdleConns=2) churn
	// connections under burst — e.g. the log-tail long-poll handler can
	// fan out 10s of concurrent CH probes per pod, and only 2 of them
	// would reuse a pooled connection. Bound the pool and keep enough
	// idle conns to absorb steady-state tail traffic without reopening.
	pool.SetMaxOpenConns(50)
	pool.SetMaxIdleConns(10)
	pool.SetConnMaxLifetime(30 * time.Minute)

	return clickhouse.Config{
		Conn: pool,
	}
}

func recordPoolMetrics(w metrics.Writer, stats sql.DBStats) {
	w.Gauge("clickhouse_pool.conns", float64(stats.MaxOpenConnections), []string{"conn_type:max"})
	w.Gauge("clickhouse_pool.conns", float64(stats.OpenConnections), []string{"conn_type:open"})
	w.Gauge("clickhouse_pool.conns", float64(stats.InUse), []string{"conn_type:in_use"})
	w.Gauge("clickhouse_pool.conns", float64(stats.Idle), []string{"conn_type:idle"})

	w.Gauge("clickhouse_pool.wait_count", float64(stats.WaitCount), nil)
	w.Gauge("clickhouse_pool.wait_duration_ms", float64(stats.WaitDuration.Milliseconds()), nil)

	w.Gauge("clickhouse_pool.conns_closed", float64(stats.MaxIdleClosed), []string{"reason:max_idle"})
	w.Gauge("clickhouse_pool.conns_closed", float64(stats.MaxIdleTimeClosed), []string{"reason:max_idle_time"})
	w.Gauge("clickhouse_pool.conns_closed", float64(stats.MaxLifetimeClosed), []string{"reason:max_lifetime"})
}
