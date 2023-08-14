package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateAppRequest struct {
	Name string `json:"name"`
}

// @BasePath /v1/apps
// Update an app
// @Summary update an app
// @Schemes
// @Description update an app
// @Param app_id path string app_id "app ID"
// @Param req body UpdateAppRequest true "Input"
// @Tags apps
// @Accept json
// @Produce json
// @Success 200 {object} app.App
// @Router /v1/apps/{app_id} [patch]
func (s *service) UpdateApp(ctx *gin.Context) {
	var req UpdateAppRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse update request: %w", err))
		return
	}

	appID := ctx.Param("app_id")
	if appID == "" {
		ctx.Error(fmt.Errorf("app_id must be passed in"))
		return
	}

	app, err := s.updateApp(ctx, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get  app%s: %w", appID, err))
		return
	}

	ctx.JSON(http.StatusOK, app)
}

func (s *service) updateApp(ctx context.Context, appID string, req *UpdateAppRequest) (*app.App, error) {
	currentApp := app.App{
		Model: app.Model{
			ID: appID,
		},
	}

	res := s.db.WithContext(ctx).Model(&currentApp).Updates(app.App{Name: req.Name})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &currentApp, nil
}
