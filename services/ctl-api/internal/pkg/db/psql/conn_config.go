package psql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"moul.io/zapgorm2"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// ConnConfig returns the connection config the service itself would use to
// reach host, resolving an IAM auth token when db_use_iam is set.
//
// Exported for preflight: a caller that builds its own password DSN would
// exercise a credential path the service may never take, so a passing check
// would say nothing about whether the service can connect.
func ConnConfig(ctx context.Context, cfg *internal.Config, host string) (*pgx.ConnConfig, error) {
	d, err := newDatabase(cfg, zapgorm2.Logger{}, nil, nil, nil, host)
	if err != nil {
		return nil, fmt.Errorf("unable to build database config: %w", err)
	}
	defer d.poolCtxCancel()

	connCfg, err := d.connCfg()
	if err != nil {
		return nil, fmt.Errorf("unable to build connection config: %w", err)
	}

	if err := d.beforeConnect(ctx, connCfg); err != nil {
		return nil, fmt.Errorf("unable to resolve database password: %w", err)
	}

	return connCfg, nil
}
