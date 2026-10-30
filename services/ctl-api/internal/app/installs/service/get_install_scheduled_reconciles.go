package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID					GetInstallScheduledReconciles
// @Summary				list scheduled reconciles for an install
// @Description			Returns scheduled reconciles for an install, newest first. Optional status filter (pending, fired,
// @Description			no-changes, cancelled, failed).
// @Param				install_id	path	string	true	"install ID"
// @Param				status		query	string	false	"filter by status"
// @Param				limit		query	int		false	"page size (default 20, max 100)"
// @Param				offset		query	int		false	"page offset"
// @Tags				installs
// @Accept				json
// @Produce			json
// @Security			APIKey
// @Security			OrgID
// @Failure			400	{object}	stderr.ErrResponse
// @Failure			401	{object}	stderr.ErrResponse
// @Failure			403	{object}	stderr.ErrResponse
// @Failure			404	{object}	stderr.ErrResponse
// @Failure			500	{object}	stderr.ErrResponse
// @Success			200	{array}		app.InstallScheduledReconcile
// @Router				/v1/installs/{install_id}/scheduled-reconciles [GET]
func (s *service) GetInstallScheduledReconciles(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	installID := ctx.Param("install_id")

	limit := 20
	if v := ctx.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 100 {
		limit = 100
	}
	offset := 0
	if v := ctx.Query("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	q := s.db.WithContext(ctx).
		Where(app.InstallScheduledReconcile{
			InstallID: installID,
			OrgID:     org.ID,
		})
	if status := ctx.Query("status"); status != "" {
		q = q.Where(app.InstallScheduledReconcile{Status: app.InstallScheduledReconcileStatus(status)})
	}

	var schedules []app.InstallScheduledReconcile
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&schedules).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install scheduled reconciles: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, schedules)
}
