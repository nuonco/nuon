package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RestartComponentRequest struct{}

//	@BasePath	/v1/components
//
// Restart an component's event loop
//
//	@Summary	restart an components event loop
//	@Schemes
//	@Description	restart component event loop
//	@Param			component_id	path	string					true	"component ID"
//	@Param			req			body	RestartComponentRequest	true	"Input"
//	@Tags			components/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{boolean}	true
//	@Router			/v1/components/{component_id}/admin-restart [POST]
func (s *service) RestartComponent(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	var req RestartComponentRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	component, err := s.getComponent(ctx, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	s.hooks.Restart(ctx, component.ID, component.App.Org.SandboxMode)
	ctx.JSON(http.StatusOK, true)
}
