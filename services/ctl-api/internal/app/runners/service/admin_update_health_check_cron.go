package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cronutil"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type AdminUpdateHealthCheckCronRequest struct {
	CronSchedule string `json:"cron_schedule" validate:"required"`
}

// @ID						AdminUpdateHealthCheckCron
// @Summary				Update the cron schedule on all process health check emitters
// @Description			Globally change the health check frequency without restarting processes
// @Param					req	body	AdminUpdateHealthCheckCronRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	object
// @Failure				400	{object}	stderr.ErrResponse
// @Router					/v1/runners/update-health-check-cron [POST]
func (s *service) AdminUpdateHealthCheckCron(ctx *gin.Context) {
	var req AdminUpdateHealthCheckCronRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := s.v.Struct(req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	sched, err := cron.ParseStandard(req.CronSchedule)
	if err != nil {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("invalid cron schedule: %w", err)))
		return
	}
	jitterWindow := validator.MinScheduleInterval(sched)

	var emitters []app.QueueEmitter
	if res := s.db.WithContext(ctx).
		Where(app.QueueEmitter{
			SignalType: "process_healthcheck",
			Mode:       app.QueueEmitterModeCron,
		}).Find(&emitters); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find health check emitters: %w", res.Error))
		return
	}

	updated := 0
	for _, em := range emitters {
		jittered := cronutil.ApplyCronJitter(em.ID, req.CronSchedule, jitterWindow)
		if em.CronSchedule == jittered {
			continue
		}

		if res := s.db.WithContext(ctx).
			Model(&app.QueueEmitter{}).
			Where("id = ?", em.ID).
			Updates(map[string]any{
				"cron_schedule": jittered,
				"jitter_window": int64(jitterWindow),
			}); res.Error != nil {
			s.l.Warn("unable to update emitter cron schedule",
				zap.String("emitter_id", em.ID),
				zap.Error(res.Error),
			)
			continue
		}
		updated++
	}

	ctx.JSON(http.StatusOK, gin.H{
		"total":   len(emitters),
		"updated": updated,
	})
}
