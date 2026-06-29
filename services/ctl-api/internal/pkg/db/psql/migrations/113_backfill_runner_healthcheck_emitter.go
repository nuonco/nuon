package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (m *Migrations) Migration113BackfillRunnerHealthcheckEmitter(ctx context.Context, db *gorm.DB) error {
	var runners []app.Runner
	if res := db.WithContext(ctx).Find(&runners); res.Error != nil {
		return fmt.Errorf("unable to list runners: %w", res.Error)
	}

	for _, runner := range runners {
		if err := m.runnersHelpers.EnsureRunnerSignalsQueue(ctx, runner.ID); err != nil {
			return fmt.Errorf("unable to ensure runner signals queue for runner %s: %w", runner.ID, err)
		}
	}

	return nil
}
