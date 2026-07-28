package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (m *Migrations) Migration113BackfillRunnerHealthcheckEmitter(ctx context.Context, db *gorm.DB) error {
	ctx, err := m.systemCtx(ctx)
	if err != nil {
		return err
	}

	var runners []app.Runner
	if res := db.WithContext(ctx).Find(&runners); res.Error != nil {
		return fmt.Errorf("unable to list runners: %w", res.Error)
	}

	// one bad runner shouldn't leave every remaining runner without a healthcheck
	var failed int
	for _, runner := range runners {
		if err := m.runnersHelpers.EnsureRunnerSignalsQueue(ctx, runner.ID); err != nil {
			failed++
			m.l.Warn("unable to ensure runner signals queue",
				zap.String("runner_id", runner.ID),
				zap.Error(err))
		}
	}
	if failed > 0 {
		m.l.Warn("finished backfilling runner healthcheck emitters with failures",
			zap.Int("failed", failed),
			zap.Int("total", len(runners)))
	}

	return nil
}
