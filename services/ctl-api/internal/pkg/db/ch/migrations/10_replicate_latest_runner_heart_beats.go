package migrations

import (
	"context"
	_ "embed"
	"strings"

	"gorm.io/gorm"
)

//go:embed 10_replicate_latest_runner_heart_beats.sql
var ReplicateLatestRunnerHeartBeats string

func (m *Migrations) Migration010ReplicateLatestRunnerHeartBeats(ctx context.Context, db *gorm.DB) error {
	for stmt := range strings.SplitSeq(ReplicateLatestRunnerHeartBeats, ";") {
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
