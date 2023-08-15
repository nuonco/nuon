package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @BasePath /v1/components
// Get latest config for a component
// @Summary get latest config for a component
// @Schemes
// @Description get latest config for a component
// @Param component_id path string component_id "component ID"
// @Tags components
// @Accept json
// @Produce json
// @Success 200 {array} app.ComponentConfigConnection
// @Router /v1/components/{component_id}/config [GET]
func (s *service) GetComponentLatestConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")
	if cmpID == "" {
		ctx.Error(fmt.Errorf("component id must be passed in"))
		return
	}

	cfgs, err := s.getComponentConfigs(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component configs: %w", err))
		return
	}
	if len(cfgs) < 1 {
		ctx.Error(fmt.Errorf("no configs defined for component"))
		return
	}

	ctx.JSON(http.StatusOK, cfgs[len(cfgs)-1])
}
