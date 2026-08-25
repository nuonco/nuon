package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type AdminCleanupOrphanedRequest struct {
	DryRun bool `json:"dry_run"`
}

type AdminCleanupOrphanedResponse struct {
	OrgsProcessed        int `json:"orgs_processed"`
	InstallsProcessed    int `json:"installs_processed"`
	AppsProcessed        int `json:"apps_processed"`
	VCSConnsProcessed    int `json:"vcs_conns_processed"`
	RunnersProcessed     int `json:"runners_processed"`
	QueuesDeleted        int `json:"queues_deleted"`
	EmittersDeleted      int `json:"emitters_deleted"`
	StrayEmittersDeleted int `json:"stray_emitters_deleted"`
	SignalsCancelled     int `json:"signals_cancelled"`
	Failed               int `json:"failed"`
}

type orphanedQueueRow struct {
	ID      string
	OwnerID string
}

// @ID						AdminCleanupOrphanedQueues
// @Summary				Delete queues and emitters orphaned by deleted orgs, installs, apps, and vcs connections
// @Description			Finds live queues whose org or owner (install, app, vcs connection, runner, or the
// @Description			runner's install) is soft- or hard-deleted, then per owning entity (in its own
// @Description			transaction) cancels pending signals, deletes the queues' emitters, and soft-deletes
// @Description			the queues. Also soft-deletes stray live emitters whose queue is already deleted or
// @Description			missing. Discovery is set-based so only entities with live queues are visited.
// @Param					req	body	AdminCleanupOrphanedRequest	true	"Input"
// @Tags					queues/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	AdminCleanupOrphanedResponse
// @Failure				400	{object}	stderr.ErrResponse
// @Router					/v1/queues/admin-cleanup-orphaned [POST]
func (s *service) AdminCleanupOrphanedQueues(ctx *gin.Context) {
	var req AdminCleanupOrphanedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	db := s.db.WithContext(ctx)

	installTable := plugins.TableName(db, app.Install{})
	appTable := plugins.TableName(db, app.App{})
	vcsConnTable := plugins.TableName(db, app.VCSConnection{})
	runnerTable := plugins.TableName(db, app.Runner{})

	type sweep struct {
		name    string
		counter *int
		query   string
		args    []any
	}

	var resp AdminCleanupOrphanedResponse

	sweeps := []sweep{
		{
			name:    "org",
			counter: &resp.OrgsProcessed,
			query: `SELECT q.id AS id, q.org_id AS owner_id FROM queues q
				LEFT JOIN orgs o ON o.id = q.org_id
				WHERE q.deleted_at = 0 AND (o.id IS NULL OR o.deleted_at != 0)`,
		},
		{
			name:    "install",
			counter: &resp.InstallsProcessed,
			query: `SELECT q.id AS id, q.owner_id AS owner_id FROM queues q
				LEFT JOIN installs i ON i.id = q.owner_id
				WHERE q.deleted_at = 0 AND q.owner_type = ? AND (i.id IS NULL OR i.deleted_at != 0)`,
			args: []any{installTable},
		},
		{
			name:    "app",
			counter: &resp.AppsProcessed,
			query: `SELECT q.id AS id, q.owner_id AS owner_id FROM queues q
				LEFT JOIN apps a ON a.id = q.owner_id
				WHERE q.deleted_at = 0 AND q.owner_type = ? AND (a.id IS NULL OR a.deleted_at != 0)`,
			args: []any{appTable},
		},
		{
			name:    "vcs_connection",
			counter: &resp.VCSConnsProcessed,
			query: `SELECT q.id AS id, q.owner_id AS owner_id FROM queues q
				LEFT JOIN vcs_connections v ON v.id = q.owner_id
				WHERE q.deleted_at = 0 AND q.owner_type = ? AND (v.id IS NULL OR v.deleted_at != 0)`,
			args: []any{vcsConnTable},
		},
		{
			name:    "runner",
			counter: &resp.RunnersProcessed,
			query: `SELECT q.id AS id, q.owner_id AS owner_id FROM queues q
				LEFT JOIN runners r ON r.id = q.owner_id
				LEFT JOIN runner_groups rg ON rg.id = r.runner_group_id
				LEFT JOIN installs i ON rg.owner_type = ? AND i.id = rg.owner_id
				WHERE q.deleted_at = 0 AND q.owner_type = ?
				AND (r.id IS NULL OR r.deleted_at != 0
					OR (rg.owner_type = ? AND (i.id IS NULL OR i.deleted_at != 0)))`,
			args: []any{installTable, runnerTable, installTable},
		},
	}

	seenQueues := make(map[string]struct{})
	for _, sw := range sweeps {
		var rows []orphanedQueueRow
		if res := db.Raw(sw.query, sw.args...).Scan(&rows); res.Error != nil {
			ctx.Error(fmt.Errorf("unable to find orphaned %s queues: %w", sw.name, res.Error))
			return
		}

		byOwner := make(map[string][]string)
		for _, row := range rows {
			if _, ok := seenQueues[row.ID]; ok {
				continue
			}
			seenQueues[row.ID] = struct{}{}
			byOwner[row.OwnerID] = append(byOwner[row.OwnerID], row.ID)
		}

		for ownerID, queueIDs := range byOwner {
			if err := s.cleanupQueues(ctx, queueIDs, req.DryRun, &resp); err != nil {
				s.l.Warn("unable to clean up orphaned queues",
					zap.String("sweep", sw.name),
					zap.String("owner_id", ownerID),
					zap.Error(err),
				)
				resp.Failed++
				continue
			}
			*sw.counter++
		}
	}

	if err := s.cleanupStrayEmitters(ctx, req.DryRun, &resp); err != nil {
		ctx.Error(err)
		return
	}

	s.l.Info("orphaned queue cleanup complete",
		zap.Bool("dry_run", req.DryRun),
		zap.Int("orgs_processed", resp.OrgsProcessed),
		zap.Int("installs_processed", resp.InstallsProcessed),
		zap.Int("apps_processed", resp.AppsProcessed),
		zap.Int("vcs_conns_processed", resp.VCSConnsProcessed),
		zap.Int("runners_processed", resp.RunnersProcessed),
		zap.Int("queues_deleted", resp.QueuesDeleted),
		zap.Int("emitters_deleted", resp.EmittersDeleted),
		zap.Int("stray_emitters_deleted", resp.StrayEmittersDeleted),
		zap.Int("signals_cancelled", resp.SignalsCancelled),
		zap.Int("failed", resp.Failed),
	)

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) cleanupQueues(ctx *gin.Context, queueIDs []string, dryRun bool, resp *AdminCleanupOrphanedResponse) error {
	var pending []app.QueueSignal
	if res := s.db.WithContext(ctx).
		Where("queue_id IN ?", queueIDs).
		Where("(status->>'status' IS NULL OR status->>'status' NOT IN ?)", terminalSignalStatuses()).
		Find(&pending); res.Error != nil {
		return fmt.Errorf("unable to list pending queue signals: %w", res.Error)
	}

	var emitterCount int64
	if res := s.db.WithContext(ctx).
		Model(&app.QueueEmitter{}).
		Where("queue_id IN ?", queueIDs).
		Count(&emitterCount); res.Error != nil {
		return fmt.Errorf("unable to count queue emitters: %w", res.Error)
	}

	if dryRun {
		resp.QueuesDeleted += len(queueIDs)
		resp.EmittersDeleted += int(emitterCount)
		resp.SignalsCancelled += len(pending)
		return nil
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, qs := range pending {
			cancelledStatus := app.CompositeStatus{
				CreatedAtTS:            time.Now().Unix(),
				Status:                 app.StatusCancelled,
				StatusHumanDescription: "cancelled by admin orphaned-queue cleanup",
				Metadata:               map[string]any{"cancelled_by": "admin-cleanup-orphaned"},
			}
			if res := tx.Model(&app.QueueSignal{}).
				Where("id = ?", qs.ID).
				Update("status", cancelledStatus); res.Error != nil {
				return fmt.Errorf("unable to cancel queue signal %s: %w", qs.ID, res.Error)
			}
		}

		if res := tx.Where("queue_id IN ?", queueIDs).Delete(&app.QueueEmitter{}); res.Error != nil {
			return fmt.Errorf("unable to delete queue emitters: %w", res.Error)
		}

		if res := tx.Where("id IN ?", queueIDs).Delete(&app.Queue{}); res.Error != nil {
			return fmt.Errorf("unable to delete queues: %w", res.Error)
		}

		return nil
	})
	if err != nil {
		return err
	}

	resp.QueuesDeleted += len(queueIDs)
	resp.EmittersDeleted += int(emitterCount)
	resp.SignalsCancelled += len(pending)
	return nil
}

// emitters left live on queues that were already deleted (or whose row is gone)
// are never reachable through the owner sweeps, so they get their own pass
func (s *service) cleanupStrayEmitters(ctx *gin.Context, dryRun bool, resp *AdminCleanupOrphanedResponse) error {
	var emitterIDs []string
	if res := s.db.WithContext(ctx).Raw(`SELECT e.id FROM queue_emitters e
		LEFT JOIN queues q ON q.id = e.queue_id
		WHERE e.deleted_at = 0 AND (q.id IS NULL OR q.deleted_at != 0)`).
		Scan(&emitterIDs); res.Error != nil {
		return fmt.Errorf("unable to find stray emitters: %w", res.Error)
	}
	if len(emitterIDs) == 0 {
		return nil
	}

	if dryRun {
		resp.StrayEmittersDeleted = len(emitterIDs)
		return nil
	}

	if res := s.db.WithContext(ctx).
		Where("id IN ?", emitterIDs).
		Delete(&app.QueueEmitter{}); res.Error != nil {
		return fmt.Errorf("unable to delete stray emitters: %w", res.Error)
	}
	resp.StrayEmittersDeleted = len(emitterIDs)
	return nil
}

func terminalSignalStatuses() []string {
	return []string{
		string(app.StatusSuccess),
		string(app.StatusCancelled),
		string(app.StatusDiscarded),
		string(app.StatusError),
	}
}
