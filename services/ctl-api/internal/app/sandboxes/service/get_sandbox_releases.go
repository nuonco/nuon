package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @ID GetSandboxReleases
// @Summary	get sandbox releases
// @Description.markdown	get_sandbox_releases.md
// @Param			sandbox_id	path	string	true	"sandbox ID"
// @Tags			sandboxes
// @Accept			json
// @Produce		json
// @Security APIKey
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.SandboxRelease
// @Router			/v1/sandboxes/{sandbox_id}/releases [get]
func (s *service) GetSandboxReleases(ctx *gin.Context) {
	sandboxID := ctx.Param("sandbox_id")

	sandbox, err := s.getSandboxReleases(ctx, sandboxID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get sandbox %s releases: %w", sandboxID, err))
		return
	}

	ctx.JSON(http.StatusOK, sandbox)
}

func (s *service) getSandboxReleases(ctx context.Context, sandboxID string) ([]app.SandboxRelease, error) {
	sandbox := app.Sandbox{}

	res := s.db.WithContext(ctx).Preload("Releases", func(db *gorm.DB) *gorm.DB {
		return db.Order("sandbox_releases.created_at DESC")
	}).First(&sandbox, "id = ?", sandboxID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox releases: %w", res.Error)
	}

	return sandbox.Releases, nil
}
