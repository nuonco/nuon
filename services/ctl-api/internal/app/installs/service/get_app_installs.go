package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@BasePath	/v1/apps/installs
//
// Get an app's installs
//
//	@Summary	get all installs for an app
//	@Schemes
//	@Description	get all installs for an org
//	@Param			app_id	path	string	true	"app ID"
//	@Tags			installs
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{array}		app.Install
//	@Router			/v1/apps/{app_id}/installs [GET]
func (s *service) GetAppInstalls(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	installs, err := s.getAppInstalls(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installs)
}

func (s *service) getAppInstalls(ctx context.Context, appID string) ([]app.Install, error) {
	currentApp := &app.App{}
	res := s.db.WithContext(ctx).
		Preload("Installs").
		Preload("Installs.SandboxRelease").
		Preload("Installs.AWSAccount").
		First(&currentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return currentApp.Installs, nil
}
