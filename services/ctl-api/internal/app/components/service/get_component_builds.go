package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/components
// Get all builds for a component
// @Summary get all builds for a component
// @Schemes
// @Description get all builds for a component
// @Param component_id path string true "component ID"
// @Tags components
// @Accept json
// @Produce json
// @Success 200 {array} app.ComponentConfigConnection
// @Router /v1/components/{component_id}/builds [GET]
func (s *service) GetComponentBuilds(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")
	if cmpID == "" {
		ctx.Error(fmt.Errorf("component id must be passed in"))
		return
	}

	component, err := s.getComponentBuilds(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component builds: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, component)
}

func (s *service) getComponentBuilds(ctx context.Context, cmpID string) ([]app.ComponentBuild, error) {
	var blds []app.ComponentBuild

	// query all builds that belong to the component id
	res := s.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("ComponentConfigConnection", "component_id = ?", cmpID).
		Find(&blds)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component build: %w", res.Error)
	}

	return blds, nil
}
