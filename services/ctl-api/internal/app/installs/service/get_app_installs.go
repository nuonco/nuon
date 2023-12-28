package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetAppInstalls
// @Summary	get all installs for an app
// @Description.markdown	get_app_installs.md
// @Param			app_id	path	string	true	"app ID"
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.Install
// @Router			/v1/apps/{app_id}/installs [GET]
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
		Preload("Installs.AppSandboxConfig").
		Preload("Installs.AWSAccount").
		First(&currentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return currentApp.Installs, nil
}
