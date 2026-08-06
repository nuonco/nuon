package helpers

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type OrgHealthcheckMigrationResult struct {
	EmittersDeleted  int      `json:"emitters_deleted"`
	EmittersCreated  int      `json:"emitters_created"`
	QueuesTerminated int      `json:"queues_terminated"`
	Errors           []string `json:"errors"`
}

// MigrateOrgToHealthcheckSweeps moves an org onto the per-org batch sweep
// emitters: ensures the sweep queue + emitters, then removes the legacy
// per-runner and per-process healthcheck cron emitters. Requires the
// org-healthcheck-sweeps feature to already be enabled. Idempotent.
func (h *Helpers) MigrateOrgToHealthcheckSweeps(ctx context.Context, orgID string) (*OrgHealthcheckMigrationResult, error) {
	enabled, err := h.featuresClient.OrgHealthcheckSweepsEnabled(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to evaluate org healthcheck sweeps flag: %w", err)
	}
	if !enabled {
		return nil, fmt.Errorf("org %s does not have the %s feature enabled", orgID, app.OrgFeatureOrgHealthcheckSweeps)
	}

	result := &OrgHealthcheckMigrationResult{Errors: []string{}}

	if err := h.EnsureOrgHealthcheckSweeps(ctx, orgID); err != nil {
		return nil, fmt.Errorf("unable to ensure org healthcheck sweeps: %w", err)
	}

	runnerIDs, err := h.orgRunnerIDs(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if len(runnerIDs) == 0 {
		return result, nil
	}

	deleted, err := h.deleteRunnerOwnedEmitters(ctx, runnerIDs, "queue_emitters.name = ?", RunnerHealthcheckEmitterName)
	if err != nil {
		return nil, fmt.Errorf("unable to delete runner healthcheck emitters: %w", err)
	}
	result.EmittersDeleted += deleted

	deleted, err = h.deleteRunnerOwnedEmitters(ctx, runnerIDs, "queue_emitters.signal_type = ?", "process_healthcheck")
	if err != nil {
		return nil, fmt.Errorf("unable to delete process healthcheck emitters: %w", err)
	}
	result.EmittersDeleted += deleted

	var cronQueues []app.Queue
	if res := h.db.WithContext(ctx).
		Select("id").
		Where("owner_type = ? AND owner_id IN ? AND name = ?", "runners", runnerIDs, RunnerHealthcheckCronsQueueName).
		Find(&cronQueues); res.Error != nil {
		return nil, fmt.Errorf("unable to list runner-healthcheck-crons queues: %w", res.Error)
	}
	for _, q := range cronQueues {
		if err := h.queueClient.Terminate(ctx, q.ID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("terminate queue %s: %v", q.ID, err))
			continue
		}
		result.QueuesTerminated++
	}

	return result, nil
}

// MigrateOrgFromHealthcheckSweeps rolls an org back to per-entity healthcheck
// cron emitters: terminates the sweep queue and recreates the per-runner and
// per-process emitters. Requires the org-healthcheck-sweeps feature to already
// be disabled. Idempotent.
func (h *Helpers) MigrateOrgFromHealthcheckSweeps(ctx context.Context, orgID string) (*OrgHealthcheckMigrationResult, error) {
	enabled, err := h.featuresClient.OrgHealthcheckSweepsEnabled(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to evaluate org healthcheck sweeps flag: %w", err)
	}
	if enabled {
		return nil, fmt.Errorf("org %s still has the %s feature enabled", orgID, app.OrgFeatureOrgHealthcheckSweeps)
	}

	result := &OrgHealthcheckMigrationResult{Errors: []string{}}

	var sweepQueue app.Queue
	res := h.db.WithContext(ctx).
		Where(app.Queue{OwnerID: orgID, Name: OrgHealthcheckCronsQueueName}).
		First(&sweepQueue)
	switch {
	case res.Error == nil:
		if err := h.queueClient.Terminate(ctx, sweepQueue.ID); err != nil {
			return nil, fmt.Errorf("unable to terminate org-healthcheck-crons queue: %w", err)
		}
		result.QueuesTerminated++
	case !errors.Is(res.Error, gorm.ErrRecordNotFound):
		return nil, fmt.Errorf("unable to find org-healthcheck-crons queue: %w", res.Error)
	}

	var runners []app.Runner
	if res := h.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Where("status NOT IN ?", []app.RunnerStatus{app.RunnerStatusDeprovisioned}).
		Find(&runners); res.Error != nil {
		return nil, fmt.Errorf("unable to list org runners: %w", res.Error)
	}

	for i := range runners {
		runner := runners[i]
		rctx := cctx.SetOrgIDContext(ctx, orgID)
		rctx = cctx.SetAccountIDContext(rctx, runner.CreatedByID)
		if err := h.EnsureRunnerHealthcheckEmitter(rctx, &runner); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("runner %s: %v", runner.ID, err))
			h.l.Warn("unable to recreate runner healthcheck emitter",
				zap.String("runner_id", runner.ID), zap.Error(err))
			continue
		}
		result.EmittersCreated++
	}

	var processes []app.RunnerProcess
	if res := h.db.WithContext(ctx).
		Select("id", "runner_id", "org_id", "created_by_id").
		Where("org_id = ?", orgID).
		Where("composite_status->>'status' IN ?", []string{
			string(app.RunnerProcessStatusActive),
			string(app.RunnerProcessStatusOffline),
		}).
		Find(&processes); res.Error != nil {
		return nil, fmt.Errorf("unable to list org runner processes: %w", res.Error)
	}

	for _, p := range processes {
		var q app.Queue
		if res := h.db.WithContext(ctx).
			Select("id").
			Where(app.Queue{OwnerID: p.RunnerID, OwnerType: "runners", Name: fmt.Sprintf("runner-process-%s", p.ID)}).
			First(&q); res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				continue
			}
			result.Errors = append(result.Errors, fmt.Sprintf("process %s queue: %v", p.ID, res.Error))
			continue
		}

		pctx := cctx.SetOrgIDContext(ctx, orgID)
		pctx = cctx.SetAccountIDContext(pctx, p.CreatedByID)
		if err := h.createProcessHealthcheckEmitter(pctx, q.ID, p.RunnerID, p.ID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("process %s: %v", p.ID, err))
			h.l.Warn("unable to recreate process healthcheck emitter",
				zap.String("process_id", p.ID), zap.Error(err))
			continue
		}
		result.EmittersCreated++
	}

	return result, nil
}

func (h *Helpers) orgRunnerIDs(ctx context.Context, orgID string) ([]string, error) {
	var runnerIDs []string
	if res := h.db.WithContext(ctx).
		Model(&app.Runner{}).
		Where("org_id = ?", orgID).
		Pluck("id", &runnerIDs); res.Error != nil {
		return nil, fmt.Errorf("unable to list org runners: %w", res.Error)
	}
	return runnerIDs, nil
}

func (h *Helpers) deleteRunnerOwnedEmitters(ctx context.Context, runnerIDs []string, cond string, condArg any) (int, error) {
	var emitters []app.QueueEmitter
	if res := h.db.WithContext(ctx).
		Joins("JOIN queues ON queues.id = queue_emitters.queue_id AND queues.deleted_at = 0").
		Where("queues.owner_type = ? AND queues.owner_id IN ?", "runners", runnerIDs).
		Where(cond, condArg).
		Find(&emitters); res.Error != nil {
		return 0, res.Error
	}

	deleted := 0
	for i := range emitters {
		em := emitters[i]
		if _, err := h.emitterClient.StopEmitter(ctx, em.ID); err != nil {
			h.l.Warn("unable to stop legacy healthcheck emitter; it will self-stop after deletion",
				zap.String("emitter_id", em.ID), zap.Error(err))
		}
		if err := h.emitterClient.DeleteEmitter(ctx, em.ID); err != nil {
			return deleted, fmt.Errorf("unable to delete emitter %s: %w", em.ID, err)
		}
		deleted++
	}
	return deleted, nil
}
