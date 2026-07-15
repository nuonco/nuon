package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	runnerSignalsQueueName   = "runner-signals"
	runnerSignalsQueueOwner  = "runners"
	runnerHealthcheckEmitter = "runner-healthcheck"
)

type BackfillRunnerHealthcheckEmittersRequest struct {
	// CursorCreatedAt / CursorID form the keyset cursor: the batch selects
	// runners ordered by (created_at, id) ascending that sort after this pair.
	// Zero values start from the beginning.
	CursorCreatedAt time.Time `json:"cursor_created_at"`
	CursorID        string    `json:"cursor_id"`
	Limit           int       `json:"limit"`
}

type BackfillRunnerHealthcheckEmittersResponse struct {
	// LastCreatedAt / LastID are the cursor of the last runner examined; feed
	// them back as the next CursorCreatedAt / CursorID. LastID is empty when
	// the batch was empty.
	LastCreatedAt time.Time `json:"last_created_at"`
	LastID        string    `json:"last_id"`
	// Examined is how many runners this batch looked at; a value < Limit means
	// the fleet is drained.
	Examined        int      `json:"examined"`
	EmittersCreated int      `json:"emitters_created"`
	AlreadyPresent  int      `json:"already_present"`
	AffectedIDs     []string `json:"affected_ids"`
	Errors          []string `json:"errors"`
}

// BackfillRunnerHealthcheckEmitters ensures the runner-healthcheck emitter
// exists for a keyset-paginated batch of runners. Idempotent per runner: it
// skips runners that already have the emitter, and creates (which also starts
// the emitter workflow) for the ones that don't. Deleted runners are excluded
// by the model's soft-delete scope.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) BackfillRunnerHealthcheckEmitters(ctx context.Context, req BackfillRunnerHealthcheckEmittersRequest) (*BackfillRunnerHealthcheckEmittersResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	var runners []app.Runner
	if res := a.db.WithContext(ctx).
		Where("(created_at, id) > (?, ?)", req.CursorCreatedAt, req.CursorID).
		Order("created_at asc, id asc").
		Limit(limit).
		Find(&runners); res.Error != nil {
		return nil, fmt.Errorf("unable to list runners: %w", res.Error)
	}

	resp := &BackfillRunnerHealthcheckEmittersResponse{
		Examined:    len(runners),
		AffectedIDs: []string{},
		Errors:      []string{},
	}

	for i := range runners {
		runner := runners[i]
		resp.LastCreatedAt = runner.CreatedAt
		resp.LastID = runner.ID

		created, err := a.ensureRunnerHealthcheckEmitter(ctx, runner)
		if err != nil {
			a.l.Warn("failed to backfill runner healthcheck emitter",
				zap.String("runner_id", runner.ID), zap.Error(err))
			resp.Errors = append(resp.Errors, fmt.Sprintf("%s: %v", runner.ID, err))
			continue
		}
		if created {
			a.l.Info("created runner healthcheck emitter",
				zap.String("runner_id", runner.ID),
				zap.String("org_id", runner.OrgID),
				zap.String("runner_group_id", runner.RunnerGroupID))
			resp.EmittersCreated++
			resp.AffectedIDs = append(resp.AffectedIDs, runner.ID)
		} else {
			resp.AlreadyPresent++
		}
	}

	a.l.Info("runner healthcheck emitter backfill batch",
		zap.Int("examined", resp.Examined),
		zap.Int("emitters_created", resp.EmittersCreated),
		zap.Int("already_present", resp.AlreadyPresent),
		zap.Strings("affected_runner_ids", resp.AffectedIDs),
		zap.Int("errors", len(resp.Errors)),
		zap.String("last_id", resp.LastID))

	return resp, nil
}

// ensureRunnerHealthcheckEmitter creates the runner-healthcheck emitter for a
// single runner if it doesn't already have one. Returns true when it created
// the emitter, false when the runner already had it.
func (a *Activities) ensureRunnerHealthcheckEmitter(ctx context.Context, runner app.Runner) (bool, error) {
	hasEmitter, err := a.runnerHasHealthcheckEmitter(ctx, runner.ID)
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

	if err := a.runnersHelpers.EnsureRunnerSignalsQueue(runnerCtx, runner.ID); err != nil {
		return false, err
	}
	return true, nil
}

// runnerHasHealthcheckEmitter reports whether the runner's runner-signals queue
// already has a runner-healthcheck emitter. Mirrors the check in
// EnsureRunnerSignalsQueue so the two agree on what "already exists" means.
func (a *Activities) runnerHasHealthcheckEmitter(ctx context.Context, runnerID string) (bool, error) {
	var q app.Queue
	res := a.db.WithContext(ctx).
		Where(&app.Queue{
			OwnerID:   runnerID,
			OwnerType: runnerSignalsQueueOwner,
			Name:      runnerSignalsQueueName,
		}).
		First(&q)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if res.Error != nil {
		return false, res.Error
	}

	var count int64
	if res := a.db.WithContext(ctx).
		Model(&app.QueueEmitter{}).
		Where("queue_id = ? AND name = ?", q.ID, runnerHealthcheckEmitter).
		Count(&count); res.Error != nil {
		return false, res.Error
	}
	return count > 0, nil
}
