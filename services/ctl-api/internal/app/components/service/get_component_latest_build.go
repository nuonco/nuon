package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/components
// Get latest build for a component
// @Summary get latest build for a component
// @Schemes
// @Description get latest build for a component
// @Param component_id path string component_id "component ID"
// @Tags components
// @Accept json
// @Produce json
// @Success 200 {object} app.ComponentBuild
// @Router /v1/components/{component_id}/latest-build [GET]
func (s *service) GetComponentLatestBuild(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")
	if cmpID == "" {
		ctx.Error(fmt.Errorf("component id must be passed in"))
		return
	}

	cfg, err := s.getComponentLatestBuild(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component builds: %w", err))
		return
	}
	if len(cfg.ComponentBuilds) < 1 {
		ctx.Error(fmt.Errorf("no builds defined for component"))
		return
	}

	ctx.JSON(http.StatusOK, cfg.ComponentBuilds[len(cfg.ComponentBuilds)-1])
}

func (s *service) getComponentLatestBuild(ctx *gin.Context, cmpID string) (*app.ComponentConfigConnection, error) {
	cmp := app.ComponentConfigConnection{}
	res := s.db.WithContext(ctx).
		Preload("Component", "id = ?", cmpID).
		Preload("ComponentBuilds").
		Preload("ComponentBuilds.VCSConnectionCommit").
		Order("created_at DESC").
		First(&cmp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get most recent component build: %w", res.Error)
	}

	return &cmp, nil
}
