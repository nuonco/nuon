package migrations

import (
	"context"
	_ "embed"
	"strings"

	"gorm.io/gorm"
)

//go:embed 06_fix_otel_log_attr_indexes.sql
var FixOtelLogAttrIndexes string

func (m *Migrations) Migration006FixOtelLogAttrIndexes(ctx context.Context, db *gorm.DB) error {
	for _, stmt := range strings.Split(FixOtelLogAttrIndexes, ";") {
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
