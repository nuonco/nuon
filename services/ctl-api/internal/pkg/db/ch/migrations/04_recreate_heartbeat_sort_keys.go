package migrations

import (
	"context"
	_ "embed"
	"strings"

	"gorm.io/gorm"
)

//go:embed 04_recreate_heartbeat_sort_keys.sql
var RecreateHeartbeatSortKeys string

func (m *Migrations) Migration004RecreateHeartbeatSortKeys(ctx context.Context, db *gorm.DB) error {
	for _, stmt := range strings.Split(RecreateHeartbeatSortKeys, ";") {
		stmt = stripSQLComments(stmt)
		if stmt == "" {
			continue
		}
		if res := db.WithContext(ctx).Exec(stmt); res.Error != nil {
			return res.Error
		}
	}

	return nil
}
