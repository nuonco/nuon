package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@BasePath	/v1/apps
//
// Create get an app
//
//	@Summary	get an app
//	@Schemes
//	@Description	get an app
//	@Param			app_id	path	string	true	"app ID"
//	@Tags			apps
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{object}	app.App
//	@Router			/v1/apps/{app_id} [get]
func (s *service) GetApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	app, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	ctx.JSON(http.StatusOK, app)
}

func (s *service) getApp(ctx context.Context, appID string) (*app.App, error) {
	app := app.App{}
	res := s.db.WithContext(ctx).
		Preload("Components").
		Preload("SandboxRelease").
		Where("name = ?", appID).
		Or("id = ?", appID).
		First(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &app, nil
}
