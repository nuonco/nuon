package routing

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

var (
	_ gorm.ConnPool       = (*ConnPool)(nil)
	_ gorm.TxBeginner     = (*ConnPool)(nil)
	_ gorm.GetDBConnector = (*ConnPool)(nil)
)

type ConnPool struct {
	primary *sql.DB
	replica *sql.DB
}

func NewConnPool(primary, replica *sql.DB) *ConnPool {
	return &ConnPool{
		primary: primary,
		replica: replica,
	}
}

func (p *ConnPool) readDB(ctx context.Context) *sql.DB {
	if p.replica != nil && DecisionFromContext(ctx) == PoolReplica {
		return p.replica
	}
	return p.primary
}

func (p *ConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.readDB(ctx).QueryContext(ctx, query, args...)
}

func (p *ConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.readDB(ctx).QueryRowContext(ctx, query, args...)
}

func (p *ConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.primary.ExecContext(ctx, query, args...)
}

func (p *ConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return p.primary.PrepareContext(ctx, query)
}

func (p *ConnPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.primary.BeginTx(ctx, opts)
}

func (p *ConnPool) GetDBConn() (*sql.DB, error) {
	return p.primary, nil
}
