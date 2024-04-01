package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

type RenderedInstaller struct {
	App        app.App              `json:"app"`
	AppInputs  app.AppInputConfig   `json:"app_inputs"`
	AppSandbox app.AppSandboxConfig `json:"app_sandbox"`

	SandboxMode bool                     `json:"sandbox_mode"`
	Metadata    app.AppInstallerMetadata `json:"metadata"`
}

// @ID RenderInstaller
// @Summary	render an installer
// @Description.markdown	render_installer.md
// @Tags installers
// @Accept			json
// @Produce		json
// @Param			installer_slug	path		string	true	"installer slug or ID"
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	RenderedInstaller
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

	if len(installer.App.AppSandboxConfigs) < 1 {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("app does not have any sandbox configs"),
			Description: "please make create at least one app sandbox config first",
		})
		return
	}
	if len(installer.App.AppRunnerConfigs) < 1 {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("app does not have any runner configs"),
			Description: "please make create at least one app runner config first",
		})
		return
	}

	if installer.App.AppSandboxConfigs[0].PublicGitVCSConfig == nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("installers currently only support sandboxes in public repos"),
			Description: "installers currently do not support custom sandboxes from connected github repos. Please make the sandbox config public, or get in touch.",
		})
		return
	}

	var inputs app.AppInputConfig
	if len(installer.App.AppInputConfigs) > 0 {
		inputs = installer.App.AppInputConfigs[0]
	}

	ctx.JSON(http.StatusCreated,
		RenderedInstaller{
			App:         installer.App,
			AppInputs:   inputs,
			AppSandbox:  installer.App.AppSandboxConfigs[0],
			SandboxMode: installer.App.Org.SandboxMode,
			Metadata:    installer.Metadata,
		})
}

func (s *service) getAppInstaller(ctx context.Context, installerID string) (*app.AppInstaller, error) {
	app := app.AppInstaller{}
	res := s.db.WithContext(ctx).
		Preload("App").
		Preload("App.Org").

		// preload sandbox
		Preload("App.AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC")
		}).
		Preload("App.AppSandboxConfigs.PublicGitVCSConfig").
		Preload("App.AppSandboxConfigs.ConnectedGithubVCSConfig").

		// preload app runner
		Preload("App.AppRunnerConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_runner_configs.created_at DESC")
		}).

		// preload app inputs
		Preload("App.AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_configs.created_at DESC")
		}).
		Preload("App.AppInputConfigs.AppInputs").

		// preload runner
		Preload("App.AppRunnerConfigs").

		// metadata
		Preload("Metadata").
		Where("slug = ?", installerID).
		Or("id = ?", installerID).
		First(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &app, nil
}
