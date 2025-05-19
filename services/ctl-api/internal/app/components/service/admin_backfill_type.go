package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
)

type BackfillRequest struct{}

// @ID						AdminBackfillRequest
// @Summary				backfill component type for components
// @Description.markdown	backfill_type.md
// @Param					component_id	path	string					true	"component ID"
// @Param					req				body	BackfillRequest	true	"Input"
// @Tags					components/admin
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/components/{component_id}/admin-backfill-type [POST]
func (s *service) AdminBackfillType(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	var req BackfillRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	component, err := s.getComponent(ctx, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	s.evClient.Send(ctx, component.ID, &signals.Signal{
		Type: signals.OperationBackillType,
	})
	ctx.JSON(http.StatusOK, true)
}
