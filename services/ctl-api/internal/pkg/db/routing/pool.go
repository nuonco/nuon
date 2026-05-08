package routing

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// Compile-time interface checks.
var (
	_ gorm.ConnPool       = (*ConnPool)(nil)
	_ gorm.TxBeginner     = (*ConnPool)(nil)
	_ gorm.GetDBConnector = (*ConnPool)(nil)
)

// ConnPool routes database operations between a primary and an optional replica.
// Read operations (QueryContext, QueryRowContext) are sent to the replica when
// the context has been marked via WithReplica and a replica is configured.
// All other operations (ExecContext, PrepareContext, BeginTx) always use the primary.
type ConnPool struct {
	primary *sql.DB
	replica *sql.DB // nil when no replica is configured
}

// NewConnPool creates a routing ConnPool. replica may be nil, in which case all
// operations are routed to primary.
func NewConnPool(primary, replica *sql.DB) *ConnPool {
	return &ConnPool{
		primary: primary,
		replica: replica,
	}
}

func (p *ConnPool) readDB(ctx context.Context) *sql.DB {
	if p.replica != nil && UseReplica(ctx) {
		return p.replica
	}
	return p.primary
}

// QueryContext routes to the replica when the context is marked for replica use.
func (p *ConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.readDB(ctx).QueryContext(ctx, query, args...)
}

// QueryRowContext routes to the replica when the context is marked for replica use.
func (p *ConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.readDB(ctx).QueryRowContext(ctx, query, args...)
}

// ExecContext always uses the primary connection.
func (p *ConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.primary.ExecContext(ctx, query, args...)
}

// PrepareContext always uses the primary connection.
func (p *ConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return p.primary.PrepareContext(ctx, query)
}

// BeginTx always uses the primary connection. Transactions must not run on the replica.
func (p *ConnPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.primary.BeginTx(ctx, opts)
}

// GetDBConn returns the primary *sql.DB for health checks and migrations.
func (p *ConnPool) GetDBConn() (*sql.DB, error) {
	return p.primary, nil
}
