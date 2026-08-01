package migrations

import (
	"context"
	_ "embed"
	"strings"

	"gorm.io/gorm"
)

//go:embed 09_install_component_resource_states_latest.sql
var InstallComponentResourceStatesLatest string

func (m *Migrations) Migration009InstallComponentResourceStatesLatest(ctx context.Context, db *gorm.DB) error {
	for stmt := range strings.SplitSeq(InstallComponentResourceStatesLatest, ";") {
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
