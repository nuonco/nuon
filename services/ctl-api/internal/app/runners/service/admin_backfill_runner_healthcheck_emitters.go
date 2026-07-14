package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	runnerSignalsQueueName   = "runner-signals"
	runnerSignalsQueueOwner  = "runners"
	runnerHealthcheckEmitter = "runner-healthcheck"

	// runnerBackfillBatchSize bounds how many runners we pull into memory at a
	// time; each batch is processed before the next is queried.
	runnerBackfillBatchSize = 20
)

type AdminBackfillRunnerHealthcheckEmittersResponse struct {
	RunnerCount       int      `json:"runner_count"`
	EmittersCreated   int      `json:"emitters_created"`
	AffectedRunnerIDs []string `json:"affected_runner_ids"`
	AlreadyPresent    int      `json:"already_present"`
	Errors            []string `json:"errors"`
}

// @ID						AdminBackfillRunnerHealthcheckEmitters
// @Summary				Backfill runner healthcheck emitters
// @Description			Ensures every runner has a runner-healthcheck emitter on its runner-signals queue, creating (and starting the emitter workflow for) the ones that predate that behavior. Idempotent: runners that already have the emitter are skipped. Returns the runner IDs that had an emitter created.
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	AdminBackfillRunnerHealthcheckEmittersResponse
// @Router					/v1/runners/backfill-health-check-emitters [POST]
func (s *service) AdminBackfillRunnerHealthcheckEmitters(ctx *gin.Context) {
	reqCtx := ctx.Request.Context()

	resp := AdminBackfillRunnerHealthcheckEmittersResponse{
		AffectedRunnerIDs: []string{},
		Errors:            []string{},
	}

	var batch []app.Runner
	res := s.db.WithContext(reqCtx).FindInBatches(&batch, runnerBackfillBatchSize, func(_ *gorm.DB, _ int) error {
		for i := range batch {
			runner := batch[i]
			resp.RunnerCount++

			created, err := s.ensureRunnerHealthcheckEmitter(reqCtx, runner)
			if err != nil {
				s.l.Warn("failed to backfill runner healthcheck emitter",
					zap.String("runner_id", runner.ID), zap.Error(err))
				resp.Errors = append(resp.Errors, fmt.Sprintf("%s: %v", runner.ID, err))
				continue
			}
			if created {
				resp.AffectedRunnerIDs = append(resp.AffectedRunnerIDs, runner.ID)
			} else {
				resp.AlreadyPresent++
			}
		}
		return nil
	})
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list runners: %w", res.Error))
		return
	}

	resp.EmittersCreated = len(resp.AffectedRunnerIDs)

	s.l.Info("backfilled runner healthcheck emitters",
		zap.Int("runner_count", resp.RunnerCount),
		zap.Int("emitters_created", resp.EmittersCreated),
		zap.Int("already_present", resp.AlreadyPresent),
		zap.Int("errors", len(resp.Errors)))

	ctx.JSON(http.StatusOK, resp)
}

// ensureRunnerHealthcheckEmitter creates the runner-healthcheck emitter for a
// single runner if it doesn't already have one. Returns true when it created the
// emitter, false when the runner already had it.
func (s *service) ensureRunnerHealthcheckEmitter(ctx context.Context, runner app.Runner) (bool, error) {
	hasEmitter, err := s.runnerHasHealthcheckEmitter(ctx, runner.ID)
	if err != nil {
		return false, err
	}
	if hasEmitter {
		return false, nil
	}

	// The runner-signals queue and healthcheck emitter are both created with
	// NOT NULL created_by_id / org_id, populated by their BeforeCreate hooks
	// from context. Set the runner's own org + creator so the inserts satisfy
	// those constraints (and the created_by_id FK), mirroring how the emitter is
	// created at runner-group creation time.
	runnerCtx := cctx.SetOrgIDContext(ctx, runner.OrgID)
	runnerCtx = cctx.SetAccountIDContext(runnerCtx, runner.CreatedByID)

	if err := s.helpers.EnsureRunnerSignalsQueue(runnerCtx, runner.ID); err != nil {
		return false, err
	}
	return true, nil
}

// runnerHasHealthcheckEmitter reports whether the runner's runner-signals queue
// already has a runner-healthcheck emitter. Mirrors the check in
// EnsureRunnerSignalsQueue so the two agree on what "already exists" means.
func (s *service) runnerHasHealthcheckEmitter(ctx context.Context, runnerID string) (bool, error) {
	var q app.Queue
	res := s.db.WithContext(ctx).
		Where(&app.Queue{
			OwnerID:   runnerID,
			OwnerType: runnerSignalsQueueOwner,
			Name:      runnerSignalsQueueName,
		}).
		First(&q)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		// No queue yet means no emitter; EnsureRunnerSignalsQueue creates both.
		return false, nil
	}
	if res.Error != nil {
		return false, res.Error
	}

	var count int64
	if res := s.db.WithContext(ctx).
		Model(&app.QueueEmitter{}).
		Where("queue_id = ? AND name = ?", q.ID, runnerHealthcheckEmitter).
		Count(&count); res.Error != nil {
		return false, res.Error
	}
	return count > 0, nil
}
