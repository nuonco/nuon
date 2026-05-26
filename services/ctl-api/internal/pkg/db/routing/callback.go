package routing

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Plugin struct {
	ACL         *TableACL
	BypassOptIn bool
	Logger      *zap.Logger
}

var _ gorm.Plugin = (*Plugin)(nil)

func (p *Plugin) Name() string { return "routing" }

func (p *Plugin) Initialize(db *gorm.DB) error {
	if err := db.Callback().Query().Before("*").Register("routing:decide", p.decide); err != nil {
		return err
	}
	if err := db.Callback().Raw().Before("*").Register("routing:decide", p.decide); err != nil {
		return err
	}
	return nil
}

func (p *Plugin) decide(tx *gorm.DB) {
	if tx.Statement == nil {
		return
	}
	ctx := tx.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}
	tx.Statement.Context = WithDecision(ctx, p.choose(ctx, tx.Statement))
}

func (p *Plugin) choose(ctx context.Context, stmt *gorm.Statement) Pool {
	pool, reason := p.route(ctx, stmt)
	if p.Logger != nil {
		p.Logger.Debug("routing decision",
			zap.String("table", stmt.Table),
			zap.String("pool", string(pool)),
			zap.String("reason", reason),
		)
	}
	return pool
}

func (p *Plugin) route(ctx context.Context, stmt *gorm.Statement) (Pool, string) {
	if IsForceReplica(ctx) {
		return PoolReplica, "force_replica"
	}
	if IsWithoutReplica(ctx) {
		return PoolPrimary, "without_replica"
	}
	if !p.ACL.AllowsReplica(stmt.Table) {
		return PoolPrimary, "acl_blocked"
	}
	for _, t := range joinTables(stmt) {
		if !p.ACL.AllowsReplica(t) {
			return PoolPrimary, "join_acl_blocked"
		}
	}
	if p.BypassOptIn {
		return PoolReplica, "bypass_opt_in"
	}
	if UseReplica(ctx) {
		return PoolReplica, "opt_in"
	}
	return PoolPrimary, "default_primary"
}

func joinTables(stmt *gorm.Statement) []string {
	if len(stmt.Joins) == 0 {
		return nil
	}
	out := make([]string, 0, len(stmt.Joins))
	for _, j := range stmt.Joins {
		out = append(out, resolveJoinTable(stmt, j.Name))
	}
	return out
}

func resolveJoinTable(stmt *gorm.Statement, name string) string {
	if stmt.Schema != nil {
		if rel, ok := stmt.Schema.Relationships.Relations[name]; ok && rel.FieldSchema != nil {
			return rel.FieldSchema.Table
		}
	}
	return name
}
