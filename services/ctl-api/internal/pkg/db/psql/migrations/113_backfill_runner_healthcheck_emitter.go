package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func (m *Migrations) Migration113BackfillRunnerHealthcheckEmitter(ctx context.Context, db *gorm.DB) error {
	var runners []app.Runner
	if res := db.WithContext(ctx).Find(&runners); res.Error != nil {
		return fmt.Errorf("unable to list runners: %w", res.Error)
	}

	var failed int
	for _, runner := range runners {
		runnerCtx := context.WithValue(ctx, keys.AccountIDCtxKey, runner.CreatedByID)

		if err := m.runnersHelpers.EnsureRunnerSignalsQueue(runnerCtx, runner.ID); err != nil {
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
