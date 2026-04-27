package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

//go:embed 04_latest_runner_heart_beats_mv_v2.sql
var LatestRunnerHeartBeatsMaterializedViewV2 string

func (m *Migrations) Migration004LatestRunnerHeartBeatsMaterializedViewV2(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec(LatestRunnerHeartBeatsMaterializedViewV2); res.Error != nil {
		return res.Error
	}

	return nil
}
