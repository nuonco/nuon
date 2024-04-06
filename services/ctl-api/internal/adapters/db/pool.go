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

func (d *database) pool() (*pgxpool.Pool, error) {
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
