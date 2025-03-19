package ch

import (
	"crypto/tls"
	"fmt"

	clickhousecore "github.com/nuonco/clickhouse-go/v2"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/gorm/clickhouse"
)

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
			"max_execution_time":               60,
			"async_insert":                     1,
			"wait_for_async_insert":            1,
			"async_insert_busy_timeout_min_ms": 200,
			"async_insert_busy_timeout_max_ms": 1000,
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

	return clickhouse.Config{
		Conn: pool,
	}
}
