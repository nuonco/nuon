package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxConnections  int32         = 10
	maxConnIdleTime time.Duration = time.Second * 15
	maxConnLifetime time.Duration = time.Minute * 5

	poolMetricsPeriod time.Duration = time.Second * 10
)

func (c *database) poolCfg() (*pgxpool.Config, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.Host,
		c.User,
		c.Password,
		c.Name,
		c.Port,
		c.SSLMode)

	connCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// configure the pool timeouts and size
	connCfg.MaxConns = maxConnections
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

	d.MetricsWriter.Gauge("gorm_pool.conns", float64(stat.TotalConns()), []string{"conn_type:total"})
	d.MetricsWriter.Gauge("gorm_pool.conns", float64(stat.AcquiredConns()), []string{"conn_type:acquired"})
	d.MetricsWriter.Gauge("gorm_pool.conns", float64(stat.ConstructingConns()), []string{"conn_type:connecting"})
	d.MetricsWriter.Gauge("gorm_pool.conns", float64(stat.IdleConns()), []string{"conn_type:idle"})
	d.MetricsWriter.Gauge("gorm_pool.conns", float64(stat.MaxConns()), []string{"conn_type:max"})

	avgAcquireDuration := stat.AcquireDuration() / time.Duration(stat.AcquireCount())
	d.MetricsWriter.Timing("gorm_pool.average_acquire_duration", avgAcquireDuration, nil)
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
