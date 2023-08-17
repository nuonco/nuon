package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/sandboxes
// Create get sandbox releases
// @Summary get sandbox releases
// @Schemes
// @Description get sandbox releases
// @Param sandbox_id path string sandbox_id "sandbox ID"
// @Tags sandboxes
// @Accept json
// @Produce json
// @Success 200 {array} app.SandboxRelease
// @Router /v1/sandboxes/{sandbox_id}/releases [get]
func (s *service) GetSandboxReleases(ctx *gin.Context) {
	sandboxID := ctx.Param("sandbox_id")

	sandbox, err := s.getSandboxReleases(ctx, sandboxID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get sandbox %s releases: %w", sandboxID, err))
		return
	}

	ctx.JSON(http.StatusOK, sandbox)
}

func (s *service) getSandboxReleases(ctx context.Context, sandboxID string) ([]*app.SandboxRelease, error) {
	var releases []*app.SandboxRelease
	sandbox := app.Sandbox{
		ID: sandboxID,
	}

	if err := s.db.WithContext(ctx).Model(&sandbox).Association("Releases").Find(&releases); err != nil {
		return nil, fmt.Errorf("unable to get sandbox releases: %w", err)
	}

	return releases, nil
}
