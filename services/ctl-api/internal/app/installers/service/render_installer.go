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
	Apps []app.App `json:"apps"`

	SandboxMode bool                  `json:"sandbox_mode"`
	Metadata    app.InstallerMetadata `json:"metadata"`
}

// @ID RenderInstaller
// @Summary	render an installer
// @Description.markdown	render_installer.md
// @Tags installers
// @Accept			json
// @Produce		json
// @Param			installer_id	path		string	true	"installer ID"
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	RenderedInstaller
// @Router			/v1/installer/{installer_id}/render [GET]
func (s *service) RenderAppInstaller(ctx *gin.Context) {
	installerID := ctx.Param("installer_id")

	installer, err := s.getInstaller(ctx, installerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app installer: %w", err))
		return
	}

	for _, installerApp := range installer.Apps {
		if len(installerApp.AppSandboxConfigs) < 1 {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("app does not have any sandbox configs"),
				Description: "please make create at least one app sandbox config first",
			})
			return
		}

		if len(installerApp.AppRunnerConfigs) < 1 {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("app does not have any runner configs"),
				Description: "please make create at least one app runner config first",
			})
			return
		}

		if installerApp.AppSandboxConfigs[0].PublicGitVCSConfig == nil {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("installers currently only support sandboxes in public repos"),
				Description: "installers currently do not support custom sandboxes from connected github repos. Please make the sandbox config public, or get in touch.",
			})
			return
		}
	}

	ctx.JSON(http.StatusCreated,
		RenderedInstaller{
			Apps:        installer.Apps,
			SandboxMode: installer.Org.SandboxMode,
			Metadata:    installer.Metadata,
		})
}

func (s *service) getInstaller(ctx context.Context, installerID string) (*app.Installer, error) {
	app := app.Installer{}
	res := s.db.WithContext(ctx).
		Preload("Apps").
		Preload("Apps.Org").

		// preload sandbox
		Preload("Apps.AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC")
		}).
		Preload("Apps.AppSandboxConfigs.PublicGitVCSConfig").
		Preload("Apps.AppSandboxConfigs.ConnectedGithubVCSConfig").

		// preload app runner
		Preload("Apps.AppRunnerConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_runner_configs.created_at DESC")
		}).

		// preload app inputs
		Preload("Apps.AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_configs.created_at DESC")
		}).
		Preload("Apps.AppInputConfigs.AppInputs").

		// metadata
		Preload("Metadata").
		First(&app, "id = ?", installerID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get installer: %w", res.Error)
	}

	return &app, nil
}
