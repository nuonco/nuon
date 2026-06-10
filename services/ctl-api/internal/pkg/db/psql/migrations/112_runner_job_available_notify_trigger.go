package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

//go:embed 112_runner_job_available_notify_trigger.sql
var RunnerJobAvailableNotifyTriggerSQL string

func (m *Migrations) Migration112RunnerJobAvailableNotifyTrigger(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec(RunnerJobAvailableNotifyTriggerSQL); res.Error != nil {
		return res.Error
	}

	return nil
}
