package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/scheduledreconcile"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type CreateInstallScheduledReconcileRequest struct {
	ScheduledAt time.Time `json:"scheduled_at" validate:"required"`
	PlanOnly    bool      `json:"plan_only"`
}

// @ID					CreateInstallScheduledReconcile
// @Summary				schedule a reconcile-app-version for a future time
// @Description			Schedules a reconcile-app-version for the install at scheduled_at (RFC3339), which must be between
// @Description			1 minute and 90 days from now. The install's pinned app config is read when the schedule fires, not
// @Description			when it is created, so a later re-pin is honored. Returns 409 if the install already has a pending
// @Description			scheduled reconcile.
// @Param				install_id	path	string									true	"install ID"
// @Param				req			body	CreateInstallScheduledReconcileRequest	true	"Input"
// @Tags				installs
// @Accept				json
// @Produce			json
// @Security			APIKey
// @Security			OrgID
// @Failure			400	{object}	stderr.ErrResponse
// @Failure			401	{object}	stderr.ErrResponse
// @Failure			403	{object}	stderr.ErrResponse
// @Failure			404	{object}	stderr.ErrResponse
// @Failure			409	{object}	stderr.ErrResponse
// @Failure			500	{object}	stderr.ErrResponse
// @Success			201	{object}	app.InstallScheduledReconcile
// @Router				/v1/installs/{install_id}/scheduled-reconciles [post]
func (s *service) CreateInstallScheduledReconcile(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateInstallScheduledReconcileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := s.v.Struct(req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if _, err := s.getInstall(ctx, installID); err != nil {
		ctx.Error(err)
		return
	}

	schedule, err := s.helpers.CreateInstallScheduledReconcileRow(ctx, installID, req.ScheduledAt, req.PlanOnly)
	if err != nil {
		ctx.Error(scheduledReconcileErrToHTTP(err))
		return
	}

	em, err := s.helpers.CreateScheduledReconcileEmitter(ctx, installID, schedule.ID, req.ScheduledAt, scheduledreconcile.SignalType, &scheduledreconcile.Signal{
		InstallID:            installID,
		ScheduledReconcileID: schedule.ID,
	})
	if err != nil {
		s.failScheduledReconcile(ctx, schedule.ID, "")
		ctx.Error(fmt.Errorf("unable to create scheduled reconcile emitter: %w", err))
		return
	}

	if err := s.helpers.SetInstallScheduledReconcileEmitterID(ctx, schedule.ID, em.ID); err != nil {
		s.failScheduledReconcile(ctx, schedule.ID, em.ID)
		ctx.Error(fmt.Errorf("unable to save scheduled reconcile emitter: %w", err))
		return
	}
	schedule.EmitterID = em.ID

	ctx.JSON(http.StatusCreated, schedule)
}

func (s *service) failScheduledReconcile(ctx *gin.Context, scheduleID, emitterID string) {
	if err := s.helpers.FailInstallScheduledReconcile(ctx, scheduleID, emitterID, "unable to create the schedule's emitter"); err != nil {
		s.l.Warn("unable to mark scheduled reconcile failed",
			zap.String("scheduled_reconcile_id", scheduleID), zap.Error(err))
	}
}

func scheduledReconcileErrToHTTP(err error) error {
	switch {
	case errors.Is(err, installhelpers.ErrScheduledReconcileAlreadyPending):
		return stderr.ErrConflict{
			Err:         err,
			Description: "The install already has a pending scheduled reconcile. Cancel it before scheduling another.",
		}
	case errors.Is(err, installhelpers.ErrScheduledReconcileTimeTooSoon), errors.Is(err, installhelpers.ErrScheduledReconcileTimeTooFar):
		return stderr.NewInvalidRequest(err)
	case errors.Is(err, installhelpers.ErrScheduledReconcileNotPending):
		return stderr.ErrConflict{
			Err:         err,
			Description: "The scheduled reconcile is not pending.",
		}
	default:
		return err
	}
}
