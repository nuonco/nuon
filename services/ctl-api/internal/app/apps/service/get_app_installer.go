package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AppInstaller struct {
	App         app.App                  `json:"app"`
	SandboxMode bool                     `json:"sandbox_mode"`
	Metadata    app.AppInstallerMetadata `json:"metadata"`
}

//	@BasePath	/v1/apps
//
// Create an app
//
//	@Summary	get an app installer
//	@Schemes
//	@Description	get an app installer
//	@Tags			apps
//	@Accept			json
//	@Produce		json
//	@Param			installer_slug	path	string				true	"installer slug or ID"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{object}	AppInstaller
//	@Router			/v1/installers/{installer_slug} [GET]
func (s *service) GetAppInstaller(ctx *gin.Context) {
	slugOrID := ctx.Param("installer_slug")
	if slugOrID == "" {
		ctx.Error(fmt.Errorf("slug or id must be set"))
		return
	}

	installer, err := s.getAppInstaller(ctx, slugOrID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app installer: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, AppInstaller{
		App:         installer.App,
		SandboxMode: installer.App.Org.SandboxMode,
		Metadata:    installer.Metadata,
	})
}

func (s *service) getAppInstaller(ctx context.Context, appID string) (*app.AppInstaller, error) {
	app := app.AppInstaller{}
	res := s.db.WithContext(ctx).
		Preload("App.Org").
		Preload("App.SandboxRelease").
		Preload("Metadata").
		Where("slug = ?", appID).
		Or("id = ?", appID).
		First(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &app, nil
}
