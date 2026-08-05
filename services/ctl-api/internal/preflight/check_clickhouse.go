package preflight

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"

	clickhousecore "github.com/ClickHouse/clickhouse-go/v2"

	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

var clickhouseCheck = Check{
	Name:        "clickhouse",
	Description: "clickhouse connectivity",

	Fields: func(cfg *internal.Config) []Field {
		return []Field{
			{Name: "clickhouse_db_host", Value: cfg.ClickhouseDBHost, Required: true},
			{Name: "clickhouse_db_port", Value: cfg.ClickhouseDBPort, Required: true},
			{Name: "clickhouse_db_name", Value: cfg.ClickhouseDBName, Required: true},
			{Name: "clickhouse_db_user", Value: cfg.ClickhouseDBUser, Required: true},
			{Name: "clickhouse_db_password", Value: cfg.ClickhouseDBPassword, Required: true, Secret: true},
			{Name: "clickhouse_db_use_tls", Value: strconv.FormatBool(cfg.ClickhouseDBUseTLS)},
			{Name: "clickhouse_db_dial_timeout", Value: cfg.ClickhouseDBDialTimeout.String()},
			{Name: "clickhouse_db_read_timeout", Value: cfg.ClickhouseDBReadTimeout.String()},
			{Name: "clickhouse_db_write_timeout", Value: cfg.ClickhouseDBWriteTimeout.String()},
		}
	},

	Probe: func(ctx context.Context, cfg *internal.Config) (string, error) {
		var tlsCfg *tls.Config
		if cfg.ClickhouseDBUseTLS {
			tlsCfg = &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // matches internal/pkg/db/ch
			}
		}

		conn, err := clickhousecore.Open(&clickhousecore.Options{
			Addr: []string{fmt.Sprintf("%s:%s", cfg.ClickhouseDBHost, cfg.ClickhouseDBPort)},
			Auth: clickhousecore.Auth{
				Database: cfg.ClickhouseDBName,
				Username: cfg.ClickhouseDBUser,
				Password: cfg.ClickhouseDBPassword,
			},
			TLS:         tlsCfg,
			DialTimeout: cfg.ClickhouseDBDialTimeout,
			ReadTimeout: cfg.ClickhouseDBReadTimeout,
		})
		if err != nil {
			return "", fmt.Errorf("open failed: %w", err)
		}
		defer conn.Close()

		if err := conn.Ping(ctx); err != nil {
			return "", fmt.Errorf("ping failed: %w", err)
		}

		var version string
		if err := conn.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
			return "", fmt.Errorf("query failed: %w", err)
		}

		return fmt.Sprintf("ClickHouse %s %s", version,
			summary("host", cfg.ClickhouseDBHost+":"+cfg.ClickhouseDBPort)), nil
	},
}
