package migrations

import (
	"context"
	_ "embed"
	"strings"

	"gorm.io/gorm"
)

//go:embed 03_add_process_id_to_heartbeats.sql
var AddProcessIDToHeartBeats string

func (m *Migrations) Migration003AddProcessIDToHeartBeats(ctx context.Context, db *gorm.DB) error {
	// ClickHouse does not support multi-statement queries in a single exec.
	// Split on semicolons and execute each statement individually.
	for _, stmt := range strings.Split(AddProcessIDToHeartBeats, ";") {
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

func stripSQLComments(s string) string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
