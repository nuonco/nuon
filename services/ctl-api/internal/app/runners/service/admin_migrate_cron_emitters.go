package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cronutil"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type AdminMigrateCronEmittersRequest struct {
	DryRun bool `json:"dry_run"`
}

type AdminMigrateCronEmittersResponse struct {
	Total            int `json:"total"`
	Updated          int `json:"updated"`
	InstallsEnqueued int `json:"installs_enqueued"`
	Skipped          int `json:"skipped"`
	Failed           int `json:"failed"`
}

// @ID						AdminMigrateCronEmitters
// @Summary				Re-jitter all cron emitters
// @Description			System emitters (process/runner/vcs health checks) are re-jittered in place from their
// @Description			canonical schedules . Install emitters (action crons, drift)
// @Description			are migrated by enqueueing an appconfig-updated reconcile per install.
// @Param					req	body	AdminMigrateCronEmittersRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	AdminMigrateCronEmittersResponse
// @Failure				400	{object}	stderr.ErrResponse
// @Router					/v1/runners/migrate-cron-emitters [POST]
func (s *service) AdminMigrateCronEmitters(ctx *gin.Context) {
	var req AdminMigrateCronEmittersRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	canonicalBySignalType := map[queuesignal.SignalType]string{
		"process_healthcheck":        helpers.ProcessHealthcheckSchedule,
		"runner_healthcheck":         helpers.RunnerHealthcheckSchedule(s.cfg.Env),
		"vcs_connection_healthcheck": vcshelpers.VCSHealthCheckSchedule,
	}
	installSignalTypes := map[queuesignal.SignalType]struct{}{
		"execute-action-workflow": {},
		"drift-check":             {},
		"drift-check-sandbox":     {},
	}

	var emitters []app.QueueEmitter
	if res := s.db.WithContext(ctx).
		Preload("Queue").
		Where(app.QueueEmitter{Mode: app.QueueEmitterModeCron}).
		Find(&emitters); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find cron emitters: %w", res.Error))
		return
	}

	resp := AdminMigrateCronEmittersResponse{Total: len(emitters)}
	installIDs := make(map[string]struct{})

	for _, em := range emitters {
		if _, ok := installSignalTypes[em.SignalType]; ok {
			if em.Queue.OwnerID != "" {
				installIDs[em.Queue.OwnerID] = struct{}{}
			}
			continue
		}

		canonical, ok := canonicalBySignalType[em.SignalType]
		if !ok {
			resp.Skipped++
			continue
		}

		jittered := cronutil.ApplyCronJitter(em.ID, canonical, cronutil.MaxJitterWindow)
		if em.CronSchedule == jittered && em.JitterWindow == cronutil.MaxJitterWindow {
			resp.Skipped++
			continue
		}
		if req.DryRun {
			resp.Updated++
			continue
		}

		if res := s.db.WithContext(ctx).
			Model(&app.QueueEmitter{}).
			Where("id = ?", em.ID).
			Updates(map[string]any{
				"cron_schedule": jittered,
				"jitter_window": int64(cronutil.MaxJitterWindow),
			}); res.Error != nil {
			s.l.Warn("unable to update emitter cron schedule",
				zap.String("emitter_id", em.ID),
				zap.Error(res.Error),
			)
			resp.Failed++
			continue
		}
		resp.Updated++
	}

	for installID := range installIDs {
		if req.DryRun {
			resp.InstallsEnqueued++
			continue
		}

		var q app.Queue
		if res := s.db.WithContext(ctx).Where(app.Queue{
			OwnerID: installID,
			Name:    installshelpers.InstallSignalsQueueName,
		}).First(&q); res.Error != nil {
			s.l.Warn("unable to find install signals queue",
				zap.String("install_id", installID),
				zap.Error(res.Error),
			)
			resp.Failed++
			continue
		}

		if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID: q.ID,
			Signal: queuesignal.NewRaw("appconfig-updated", map[string]any{
				"install_id": installID,
			}),
		}); err != nil {
			s.l.Warn("unable to enqueue install reconcile signal",
				zap.String("install_id", installID),
				zap.Error(err),
			)
			resp.Failed++
			continue
		}
		resp.InstallsEnqueued++
	}

	ctx.JSON(http.StatusOK, resp)
}
