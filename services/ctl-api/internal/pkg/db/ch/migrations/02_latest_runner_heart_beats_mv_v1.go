package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

//go:embed 02_latest_runner_heart_beats_mv_v1.sql
var LatestRunnerHeartBeatsMaterializedViewV1 string

func (m *Migrations) Migration002LatestRunnerHeartBeatsMaterializedViewV1(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec(LatestRunnerHeartBeatsMaterializedViewV1); res.Error != nil {
		return res.Error
	}

	return nil
}
