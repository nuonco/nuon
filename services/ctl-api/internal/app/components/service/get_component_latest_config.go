package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/components
// Get latest config for a component
// @Summary get latest config for a component
// @Schemes
// @Description get latest config for a component
// @Param component_id path string true "component ID"
// @Tags components
// @Accept json
// @Produce json
// @Success 200 {array} app.ComponentConfigConnection
// @Router /v1/components/{component_id}/config [GET]
func (s *service) GetComponentLatestConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

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

func (s *service) getComponentLatestConfig(ctx *gin.Context, cmpID string) (*app.ComponentConfigConnection, error) {
	cmp := app.ComponentConfigConnection{}
	res := s.db.WithContext(ctx).
		Preload("Component", "id = ?", cmpID).
		Order("created_at DESC").
		First(&cmp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get most recent component config: %w", res.Error)
	}

	return &cmp, nil
}
