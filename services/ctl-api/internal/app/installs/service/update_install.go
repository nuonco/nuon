package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateInstallRequest struct {
	Name string `json:"name"`
}

// @BasePath /v1/installs
// Update an install
// @Summary update an install
// @Schemes
// @Description update an install
// @Param install_id path string app_id "app ID"
// @Param req body UpdateInstallRequest true "Input"
// @Tags installs
// @Accept json
// @Produce json
// @Success 201 {object} app.Install
// @Router /v1/{install_id} [PATCH]
func (s *service) UpdateInstall(ctx *gin.Context) {
	var req UpdateInstallRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse update request: %w", err))
		return
	}

	installID := ctx.Param("install_id")
	if installID == "" {
		ctx.Error(fmt.Errorf("install_id must be passed in"))
		return
	}

	install, err := s.updateInstall(ctx, installID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get  app%s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) updateInstall(ctx context.Context, installID string, req *UpdateInstallRequest) (*app.Install, error) {
	currentInstall := app.Install{
		ID: installID,
	}

	res := s.db.WithContext(ctx).Model(&currentInstall).Updates(app.Install{Name: req.Name})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &currentInstall, nil
}
