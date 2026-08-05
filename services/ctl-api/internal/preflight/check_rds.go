package preflight

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"

	internal "github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
)

var rdsCheck = Check{
	Name:        "rds",
	Description: "postgres connectivity",

	Fields: func(cfg *internal.Config) []Field {
		return []Field{
			{Name: "db_host", Value: cfg.DBHost, Required: true},
			{Name: "db_port", Value: cfg.DBPort, Required: true},
			{Name: "db_name", Value: cfg.DBName, Required: true},
			{Name: "db_user", Value: cfg.DBUser, Required: true},
			{Name: "db_ssl_mode", Value: cfg.DBSSLMode, Required: true},
			{Name: "db_use_iam", Value: strconv.FormatBool(cfg.DBUseIAM)},
			// Exactly one credential path applies: an IAM token is minted per
			// connection, otherwise the static password is used.
			{Name: "db_password", Value: cfg.DBPassword, Required: !cfg.DBUseIAM, Secret: true},
			{Name: "db_region", Value: cfg.DBRegion, Required: cfg.DBUseIAM},
			{Name: "cloud_provider", Value: cfg.CloudProvider},
		}
	},

	Probe: func(ctx context.Context, cfg *internal.Config) (string, error) {
		connCfg, err := psql.ConnConfig(ctx, cfg, cfg.DBHost)
		if err != nil {
			return "", err
		}

		conn, err := pgx.ConnectConfig(ctx, connCfg)
		if err != nil {
			return "", fmt.Errorf("connection failed: %w", err)
		}
		defer conn.Close(ctx)

		var version string
		if err := conn.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
			return "", fmt.Errorf("query failed: %w", err)
		}

		auth := "password"
		if cfg.DBUseIAM {
			auth = "iam"
		}

		return fmt.Sprintf("%s %s", truncate(version, 40),
			summary("host", cfg.DBHost+":"+cfg.DBPort, "auth", auth)), nil
	},
}
