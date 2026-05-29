package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxConnIdleTime time.Duration = time.Second * 15
	maxConnLifetime time.Duration = time.Minute * 5

	poolMetricsPeriod time.Duration = time.Second * 10
)

func (c *database) poolCfg() (*pgxpool.Config, error) {
	// Omit password= from the DSN when using IAM auth (PasswordFn set) — an empty
	// password= value followed immediately by dbname= can confuse pgconn's keyword-value
	// parser, causing the database field to be silently dropped. The password is injected
	// per-connection via beforeConnect instead.
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

	connCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// configure the pool timeouts and size
	connCfg.MaxConns = c.MaxConnections
	connCfg.MaxConnIdleTime = maxConnIdleTime
	connCfg.MaxConnLifetime = maxConnLifetime

	// configure the pool to use our password function to get the RDS password
	connCfg.BeforeConnect = c.beforeConnect

	return connCfg, nil
}

// beforeConnect is used to create connections using a password function, such as using AWS RDS to get a one off
// password
func (d *database) beforeConnect(ctx context.Context, connCfg *pgx.ConnConfig) error {
	if d.PasswordFn == nil {
		return nil
	}

	password, err := d.PasswordFn(ctx, *d)
	if err != nil {
		return err
	}

	c := connCfg.Config
	c.Password = password
	connCfg.Config = c
	connCfg.Password = password
	return nil
}

func (d *database) createPool() (*pgxpool.Pool, error) {
	connCfg, err := d.poolCfg()
	if err != nil {
		return nil, fmt.Errorf("unable to create database connection config: %w", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, connCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create pool: %w", err)
	}
	return pool, nil
}

func (d *database) recordPoolMetrics() {
	stat := d.pool.Stat()

	role := d.Role
	if role == "" {
		role = "primary"
	}
	roleTag := "pool:" + role

	d.MetricsWriter.Gauge("gorm_pool.conns", float64(stat.TotalConns()), []string{"conn_type:total", roleTag})
	d.MetricsWriter.Gauge("gorm_pool.conns", float64(stat.AcquiredConns()), []string{"conn_type:acquired", roleTag})
	d.MetricsWriter.Gauge("gorm_pool.conns", float64(stat.ConstructingConns()), []string{"conn_type:connecting", roleTag})
	d.MetricsWriter.Gauge("gorm_pool.conns", float64(stat.IdleConns()), []string{"conn_type:idle", roleTag})
	d.MetricsWriter.Gauge("gorm_pool.conns", float64(stat.MaxConns()), []string{"conn_type:max", roleTag})

	// empty = waited for a conn; canceled = caller ctx expired while waiting (pool too small)
	d.MetricsWriter.Gauge("gorm_pool.acquire", float64(stat.AcquireCount()), []string{"acquire_type:total", roleTag})
	d.MetricsWriter.Gauge("gorm_pool.acquire", float64(stat.EmptyAcquireCount()), []string{"acquire_type:empty", roleTag})
	d.MetricsWriter.Gauge("gorm_pool.acquire", float64(stat.CanceledAcquireCount()), []string{"acquire_type:canceled", roleTag})
	d.MetricsWriter.Gauge("gorm_pool.acquire_duration_ms", float64(stat.AcquireDuration().Milliseconds()), []string{roleTag})

	d.MetricsWriter.Gauge("gorm_pool.conns_destroyed", float64(stat.NewConnsCount()), []string{"reason:new", roleTag})
	d.MetricsWriter.Gauge("gorm_pool.conns_destroyed", float64(stat.MaxIdleDestroyCount()), []string{"reason:idle", roleTag})
	d.MetricsWriter.Gauge("gorm_pool.conns_destroyed", float64(stat.MaxLifetimeDestroyCount()), []string{"reason:lifetime", roleTag})
}

func (d *database) startPoolBackgroundJob() {
	ticker := time.NewTicker(poolMetricsPeriod)
	go func() {
		for {
			select {
			case <-d.poolCtx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				d.recordPoolMetrics()
			}
		}
	}()
}

func (d *database) stopPoolBackgroundJob() {
	d.poolCtxCancel()
}
