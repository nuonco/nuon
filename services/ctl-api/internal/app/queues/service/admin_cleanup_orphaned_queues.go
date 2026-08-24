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
	OrgsProcessed     int `json:"orgs_processed"`
	InstallsProcessed int `json:"installs_processed"`
	AppsProcessed     int `json:"apps_processed"`
	VCSConnsProcessed int `json:"vcs_conns_processed"`
	QueuesDeleted     int `json:"queues_deleted"`
	EmittersDeleted   int `json:"emitters_deleted"`
	SignalsCancelled  int `json:"signals_cancelled"`
	Failed            int `json:"failed"`
}

// @ID						AdminCleanupOrphanedQueues
// @Summary				Delete queues and emitters orphaned by forgotten installs and deleted orgs and apps
// @Description			Iterates soft-deleted orgs, forgotten installs, deleted apps (including apps of
// @Description			deleted orgs), and deleted VCS connections, and for each one (in its own transaction)
// @Description			cancels pending signals on its live queues, deletes the queues' emitters, and
// @Description			soft-deletes the queues. Install cleanup also covers queues owned by runners in the
// @Description			install's runner groups.
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

	var deletedOrgIDs []string
	if res := db.Unscoped().Model(&app.Org{}).
		Where("deleted_at != 0").
		Pluck("id", &deletedOrgIDs); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list deleted orgs: %w", res.Error))
		return
	}

	var forgottenInstallIDs []string
	if res := db.Unscoped().Model(&app.Install{}).
		Where("deleted_at != 0").
		Pluck("id", &forgottenInstallIDs); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list forgotten installs: %w", res.Error))
		return
	}

	var deletedAppIDs []string
	if res := db.Unscoped().Model(&app.App{}).
		Where("deleted_at != 0").
		Or("org_id IN (?)", db.Session(&gorm.Session{NewDB: true}).Unscoped().
			Model(&app.Org{}).Select("id").Where("deleted_at != 0")).
		Pluck("id", &deletedAppIDs); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list deleted apps: %w", res.Error))
		return
	}

	var resp AdminCleanupOrphanedResponse

	for _, orgID := range deletedOrgIDs {
		var queueIDs []string
		if res := db.Model(&app.Queue{}).
			Where(app.Queue{OrgID: &orgID}).
			Pluck("id", &queueIDs); res.Error != nil {
			s.l.Warn("unable to list deleted-org queues", zap.String("org_id", orgID), zap.Error(res.Error))
			resp.Failed++
			continue
		}
		if len(queueIDs) == 0 {
			continue
		}

		if err := s.cleanupQueues(ctx, queueIDs, req.DryRun, &resp); err != nil {
			s.l.Warn("unable to clean up deleted-org queues", zap.String("org_id", orgID), zap.Error(err))
			resp.Failed++
			continue
		}
		resp.OrgsProcessed++
	}

	installTable := plugins.TableName(db, app.Install{})
	runnerTable := plugins.TableName(db, app.Runner{})

	for _, installID := range forgottenInstallIDs {
		var queueIDs []string
		if res := db.Model(&app.Queue{}).
			Where(app.Queue{OwnerID: installID, OwnerType: installTable}).
			Pluck("id", &queueIDs); res.Error != nil {
			s.l.Warn("unable to list install queues", zap.String("install_id", installID), zap.Error(res.Error))
			resp.Failed++
			continue
		}

		var runnerGroupIDs []string
		if res := db.Unscoped().Model(&app.RunnerGroup{}).
			Where(app.RunnerGroup{OwnerID: installID, OwnerType: installTable}).
			Pluck("id", &runnerGroupIDs); res.Error != nil {
			s.l.Warn("unable to list install runner groups", zap.String("install_id", installID), zap.Error(res.Error))
			resp.Failed++
			continue
		}

		if len(runnerGroupIDs) > 0 {
			var runnerIDs []string
			if res := db.Unscoped().Model(&app.Runner{}).
				Where("runner_group_id IN ?", runnerGroupIDs).
				Pluck("id", &runnerIDs); res.Error != nil {
				s.l.Warn("unable to list install runners", zap.String("install_id", installID), zap.Error(res.Error))
				resp.Failed++
				continue
			}

			if len(runnerIDs) > 0 {
				var runnerQueueIDs []string
				if res := db.Model(&app.Queue{}).
					Where(app.Queue{OwnerType: runnerTable}).
					Where("owner_id IN ?", runnerIDs).
					Pluck("id", &runnerQueueIDs); res.Error != nil {
					s.l.Warn("unable to list runner queues", zap.String("install_id", installID), zap.Error(res.Error))
					resp.Failed++
					continue
				}
				queueIDs = append(queueIDs, runnerQueueIDs...)
			}
		}

		if len(queueIDs) == 0 {
			continue
		}

		if err := s.cleanupQueues(ctx, queueIDs, req.DryRun, &resp); err != nil {
			s.l.Warn("unable to clean up install queues", zap.String("install_id", installID), zap.Error(err))
			resp.Failed++
			continue
		}
		resp.InstallsProcessed++
	}

	var deletedVCSConnIDs []string
	if res := db.Unscoped().Model(&app.VCSConnection{}).
		Where("deleted_at != 0").
		Pluck("id", &deletedVCSConnIDs); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list deleted vcs connections: %w", res.Error))
		return
	}

	vcsConnTable := plugins.TableName(db, app.VCSConnection{})

	for _, vcsConnID := range deletedVCSConnIDs {
		var queueIDs []string
		if res := db.Model(&app.Queue{}).
			Where(app.Queue{OwnerID: vcsConnID, OwnerType: vcsConnTable}).
			Pluck("id", &queueIDs); res.Error != nil {
			s.l.Warn("unable to list vcs connection queues", zap.String("vcs_connection_id", vcsConnID), zap.Error(res.Error))
			resp.Failed++
			continue
		}
		if len(queueIDs) == 0 {
			continue
		}

		if err := s.cleanupQueues(ctx, queueIDs, req.DryRun, &resp); err != nil {
			s.l.Warn("unable to clean up vcs connection queues", zap.String("vcs_connection_id", vcsConnID), zap.Error(err))
			resp.Failed++
			continue
		}
		resp.VCSConnsProcessed++
	}

	appTable := plugins.TableName(db, app.App{})

	for _, appID := range deletedAppIDs {
		var queueIDs []string
		if res := db.Model(&app.Queue{}).
			Where(app.Queue{OwnerID: appID, OwnerType: appTable}).
			Pluck("id", &queueIDs); res.Error != nil {
			s.l.Warn("unable to list app queues", zap.String("app_id", appID), zap.Error(res.Error))
			resp.Failed++
			continue
		}
		if len(queueIDs) == 0 {
			continue
		}

		if err := s.cleanupQueues(ctx, queueIDs, req.DryRun, &resp); err != nil {
			s.l.Warn("unable to clean up app queues", zap.String("app_id", appID), zap.Error(err))
			resp.Failed++
			continue
		}
		resp.AppsProcessed++
	}

	s.l.Info("orphaned queue cleanup complete",
		zap.Bool("dry_run", req.DryRun),
		zap.Int("orgs_processed", resp.OrgsProcessed),
		zap.Int("installs_processed", resp.InstallsProcessed),
		zap.Int("apps_processed", resp.AppsProcessed),
		zap.Int("vcs_conns_processed", resp.VCSConnsProcessed),
		zap.Int("queues_deleted", resp.QueuesDeleted),
		zap.Int("emitters_deleted", resp.EmittersDeleted),
		zap.Int("signals_cancelled", resp.SignalsCancelled),
		zap.Int("failed", resp.Failed),
	)

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) cleanupQueues(ctx *gin.Context, queueIDs []string, dryRun bool, resp *AdminCleanupOrphanedResponse) error {
	var signals []app.QueueSignal
	if res := s.db.WithContext(ctx).
		Where("queue_id IN ?", queueIDs).
		Find(&signals); res.Error != nil {
		return fmt.Errorf("unable to list queue signals: %w", res.Error)
	}

	pending := make([]app.QueueSignal, 0, len(signals))
	for _, qs := range signals {
		if !isTerminalSignalStatus(qs.Status.Status) {
			pending = append(pending, qs)
		}
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

func isTerminalSignalStatus(s app.Status) bool {
	switch s {
	case app.StatusSuccess, app.StatusCancelled, app.StatusDiscarded, app.StatusError:
		return true
	}
	return false
}
