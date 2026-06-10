package psql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// NewPrimaryListenerConn opens a single dedicated, non-pooled connection to the
// PRIMARY database, suitable for LISTEN/NOTIFY. The caller owns the connection
// and must Close it.
//
// This deliberately bypasses the pgxpool: a LISTEN binds to a session and pins
// a connection for its lifetime, which is the wrong thing to take from a shared
// pool. Because RDS IAM auth tokens rotate (~15m) and pooled connections are
// recycled on a 5m lifetime, callers should periodically tear this connection
// down and re-open via this function rather than holding it open indefinitely —
// each reopen re-runs the IAM token fetch below.
func NewPrimaryListenerConn(ctx context.Context, cfg *internal.Config) (*pgx.Conn, error) {
	d := &database{
		Host:    cfg.DBHost,
		User:    cfg.DBUser,
		Name:    cfg.DBName,
		Port:    cfg.DBPort,
		SSLMode: cfg.DBSSLMode,
		Region:  cfg.DBRegion,
	}
	switch {
	case cfg.DBPassword != "":
		d.Password = cfg.DBPassword
	case cfg.DBUseIAM && cfg.IsGCP():
		d.PasswordFn = FetchGcpCloudSqlPassword
	case cfg.DBUseIAM:
		d.PasswordFn = FetchIamTokenPassword
	}

	connCfg, err := d.connCfg()
	if err != nil {
		return nil, fmt.Errorf("unable to build listener conn config: %w", err)
	}

	// connCfg() omits password from the DSN when a PasswordFn is set (mirrors
	// the pool's beforeConnect hook); inject the freshly-fetched token here.
	if d.PasswordFn != nil {
		pw, err := d.PasswordFn(ctx, *d)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch listener db password: %w", err)
		}
		connCfg.Password = pw
	}

	conn, err := pgx.ConnectConfig(ctx, connCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to open listener connection: %w", err)
	}

	return conn, nil
}
