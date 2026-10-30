package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID					CancelInstallScheduledReconcile
// @Summary				cancel a pending scheduled reconcile
// @Description			Cancels a pending scheduled reconcile for an install and deletes its backing emitter. Returns 409
// @Description			if the schedule is not pending.
// @Param				install_id				path	string	true	"install ID"
// @Param				scheduled_reconcile_id	path	string	true	"scheduled reconcile ID"
// @Tags				installs
// @Accept				json
// @Produce			json
// @Security			APIKey
// @Security			OrgID
// @Failure			401	{object}	stderr.ErrResponse
// @Failure			403	{object}	stderr.ErrResponse
// @Failure			404	{object}	stderr.ErrResponse
// @Failure			409	{object}	stderr.ErrResponse
// @Failure			500	{object}	stderr.ErrResponse
// @Success			200	{object}	app.InstallScheduledReconcile
// @Router				/v1/installs/{install_id}/scheduled-reconciles/{scheduled_reconcile_id} [DELETE]
func (s *service) CancelInstallScheduledReconcile(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	installID := ctx.Param("install_id")
	scheduleID := ctx.Param("scheduled_reconcile_id")

	var schedule app.InstallScheduledReconcile
	if err := s.db.WithContext(ctx).
		Where(app.InstallScheduledReconcile{
			ID:        scheduleID,
			InstallID: installID,
			OrgID:     org.ID,
		}).
		First(&schedule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{Err: fmt.Errorf("scheduled reconcile %s not found", scheduleID)})
			return
		}
		ctx.Error(fmt.Errorf("unable to get scheduled reconcile: %w", err))
		return
	}

	if err := s.helpers.CancelInstallScheduledReconcile(ctx, &schedule, "cancelled by user"); err != nil {
		if errors.Is(err, installhelpers.ErrScheduledReconcileNotPending) {
			ctx.Error(stderr.ErrConflict{
				Err:         err,
				Description: "The scheduled reconcile is not pending.",
			})
			return
		}
		ctx.Error(fmt.Errorf("unable to cancel scheduled reconcile: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}
