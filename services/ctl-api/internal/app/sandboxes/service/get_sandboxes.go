package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetSandboxes
// @Summary	get all sandboxes
// @Description.markdown	get_sandboxes.md
// @Tags			sandboxes
// @Accept			json
// @Produce		json
// @Success		200	{array}	app.Sandbox
// @Router			/v1/sandboxes [get]
func (s *service) GetSandboxes(ctx *gin.Context) {
	sandboxes, err := s.getSandboxes(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get sandboxes: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, sandboxes)
}

func (s *service) getSandboxes(ctx *gin.Context) ([]*app.Sandbox, error) {
	var sandboxes []*app.Sandbox

	res := s.db.WithContext(ctx).Preload("Releases").Find(&sandboxes)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all sandboxes: %w", res.Error)
	}

	return sandboxes, nil
}
