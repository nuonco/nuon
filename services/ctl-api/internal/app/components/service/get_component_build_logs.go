package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type BuildLog struct{}

// @BasePath /v1/installs
// Get install build logs
// @Summary get install build logs
// @Schemes
// @Description get install build logs
// @Param component_id path string true "component ID"
// @Param build_id path string true "build ID"
// @Tags components
// @Accept json
// @Produce json
// @Success 200 {object} []BuildLog
// @Router /v1/components/{component_id}/builds/{build_id}/logs [get]
func (s *service) GetComponentBuildLogs(ctx *gin.Context) {
	ctx.Error(fmt.Errorf("not yet implemented"))
}
