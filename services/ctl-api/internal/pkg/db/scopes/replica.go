package scopes

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
)

func WithReplica(db *gorm.DB) *gorm.DB {
	db.Statement.Context = routing.WithReplica(stmtContext(db))
	return db
}

func ForceReplica(db *gorm.DB) *gorm.DB {
	db.Statement.Context = routing.WithForceReplica(stmtContext(db))
	return db
}

func WithoutReplica(db *gorm.DB) *gorm.DB {
	db.Statement.Context = routing.WithoutReplica(stmtContext(db))
	return db
}

func stmtContext(db *gorm.DB) context.Context {
	if db.Statement != nil && db.Statement.Context != nil {
		return db.Statement.Context
	}
	return context.Background()
}
