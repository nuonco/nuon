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

	// one bad runner shouldn't leave every remaining runner without a healthcheck
	var failed int
	for _, runner := range runners {
		// the queue and emitter are created through BeforeCreate hooks that read the acting
		// account off the context, which a migration has none of. the runner's own
		// created_by_id is not null with an FK to accounts, so it is always a usable id.
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
