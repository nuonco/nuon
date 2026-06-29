package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						CancelAppComponentBuild
// @Summary				cancel component build
// @Description.markdown	cancel_component_build.md
// @Param					app_id			path	string	true	"app ID"
// @Param					component_id	path	string	true	"component ID"
// @Param					build_id		path	string	true	"build ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				202	{object}	app.ComponentBuild
// @Router					/v1/apps/{app_id}/components/{component_id}/builds/{build_id}/cancel [POST]
func (s *service) CancelAppComponentBuild(ctx *gin.Context) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	bldID := ctx.Param("build_id")

	var bld app.ComponentBuild
	if err := s.db.WithContext(ctx).
		Preload("QueueSignal").
		First(&bld, "id = ? AND org_id = ?", bldID, orgID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get build: %w", err))
		return
	}

	if bld.QueueSignal == nil {
		ctx.Error(fmt.Errorf("build has no queue signal"))
		return
	}

	if _, err := s.queueClient.CancelSignal(ctx, bld.QueueSignal.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to cancel build signal: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, bld)
}
