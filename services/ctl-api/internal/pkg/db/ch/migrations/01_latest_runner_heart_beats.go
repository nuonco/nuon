package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

//go:embed 01_latest_runner_heart_beats.sql
var LatestRunnerHeartBeats string

func (m *Migrations) Migration001LatestRunnerHeartBeats(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec(LatestRunnerHeartBeats); res.Error != nil {
		return res.Error
	}

	return nil
}
