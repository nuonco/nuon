package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type AppInstaller struct {
	App         app.App                  `json:"app"`
	SandboxMode bool                     `json:"sandbox_mode"`
	Metadata    app.AppInstallerMetadata `json:"metadata"`
}

// @ID RenderAppInstaller
// @Summary	render an app installer
// @Description.markdown	render_app_installer.md
// @Tags			apps
// @Accept			json
// @Produce		json
// @Param			installer_slug	path		string	true	"installer slug or ID"
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	AppInstaller
// @Router			/v1/installer/{installer_slug}/render [GET]
func (s *service) RenderAppInstaller(ctx *gin.Context) {
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

func (s *service) getAppInstaller(ctx context.Context, installerID string) (*app.AppInstaller, error) {
	app := app.AppInstaller{}
	res := s.db.WithContext(ctx).
		Preload("App").
		Preload("App.Org").
		Preload("App.AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC")
		}).
		Preload("App.AppSandboxConfigs.PublicGitVCSConfig").
		Preload("App.AppSandboxConfigs.ConnectedGithubVCSConfig").
		Preload("App.AppSandboxConfigs.SandboxRelease").
		Preload("App.AppSandboxConfigs.SandboxRelease.Sandbox").
		Preload("Metadata").
		Where("slug = ?", installerID).
		Or("id = ?", installerID).
		First(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &app, nil
}
