package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/sandboxes
// Create get a sandbox
// @Summary get a sandbox
// @Schemes
// @Description get a sandbox
// @Param sandbox_id path string true "sandbox ID"
// @Tags sandboxes
// @Accept json
// @Produce json
// @Success 200 {object} app.Sandbox
// @Router /v1/sandboxes/{sandbox_id} [get]
func (s *service) GetSandbox(ctx *gin.Context) {
	sandboxID := ctx.Param("sandbox_id")

	sandbox, err := s.getSandbox(ctx, sandboxID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get sandbox %s: %w", sandboxID, err))
		return
	}

	ctx.JSON(http.StatusOK, sandbox)
}

func (s *service) getSandbox(ctx context.Context, sandboxID string) (*app.Sandbox, error) {
	sandbox := app.Sandbox{}
	res := s.db.WithContext(ctx).
		Preload("Releases").
		Where("name = ?", sandboxID).
		Or("id = ?", sandboxID).
		First(&sandbox)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox: %w", res.Error)
	}

	return &sandbox, nil
}
