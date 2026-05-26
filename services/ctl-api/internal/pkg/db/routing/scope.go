package routing

import (
	"context"

	"gorm.io/gorm"
)

func Replica(db *gorm.DB) *gorm.DB {
	return db.WithContext(WithReplica(stmtContext(db)))
}

func ForceReplica(db *gorm.DB) *gorm.DB {
	return db.WithContext(WithForceReplica(stmtContext(db)))
}

func Primary(db *gorm.DB) *gorm.DB {
	return db.WithContext(WithoutReplica(stmtContext(db)))
}

func stmtContext(db *gorm.DB) context.Context {
	if db.Statement != nil && db.Statement.Context != nil {
		return db.Statement.Context
	}
	return context.Background()
}
