package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
)

type AdminDeleteComponentRequest struct{}

// @ID AdminDeleteComponent
// @Summary	delete a component
// @Description.markdown delete_component.md
// @Param			component_id	path	string						true	"component ID"
// @Param			req				body	AdminDeleteComponentRequest	true	"Input"
// @Tags			components/admin
// @Accept			json
// @Produce		json
// @Success		200	{boolean}	true
// @Router			/v1/components/{component_id}/admin-delete [POST]
func (s *service) AdminDeleteComponent(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	s.evClient.Send(ctx, componentID, &signals.Signal{
		Type: signals.OperationDeleted,
	})
	ctx.JSON(http.StatusOK, true)
}
