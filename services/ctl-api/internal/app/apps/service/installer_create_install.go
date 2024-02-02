package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

type InstallerCreateInstallRequest struct {
	Name string `json:"name" validate:"required"`

	AWSAccount struct {
		Region     string `json:"region"`
		IAMRoleARN string `json:"iam_role_arn" validate:"required"`
	} `json:"aws_account" validate:"required"`
}

func (c *InstallerCreateInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID InstallerCreateInstall
// @Summary	create an app install from an installer
// @Description.markdown installer_create_install.md
// @Param			req	body	InstallerCreateInstallRequest	true	"Input"
// @Tags			installs
// @Accept			json
// @Produce		json
// @Param			installer_slug	path		string	true	"installer slug or ID"
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.Install
// @Router			/v1/installer/{installer_slug}/installs [post]
func (s *service) CreateInstallerInstall(ctx *gin.Context) {
	var req InstallerCreateInstallRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	installerSlug := ctx.Param("installer_slug")
	installer, err := s.getAppInstaller(ctx, installerSlug)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installer: %w", err))
		return
	}

	cctx := context.WithValue(ctx, "org_id", installer.App.OrgID)
	cctx = context.WithValue(cctx, "user_id", installer.ID)
	install, err := s.createInstall(cctx, installer.App.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	s.installHooks.Created(cctx, install.ID, installer.App.Org.SandboxMode)
	ctx.JSON(http.StatusCreated, install)
}

func (s *service) createInstall(ctx context.Context, appID string, req *InstallerCreateInstallRequest) (*app.Install, error) {
	parentApp := app.App{}
	res := s.db.WithContext(ctx).Preload("Components").
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC")
		}).
		Preload("AppRunnerConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_runner_configs.created_at DESC")
		}).First(&parentApp, "id = ?", appID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	if len(parentApp.AppSandboxConfigs) < 1 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("app does not have any sandbox configs"),
			Description: "please make create at least one app sandbox config first",
		}
	}
	if len(parentApp.AppRunnerConfigs) < 1 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("app does not have any runner configs"),
			Description: "please make create at least one app runner config first",
		}
	}

	installCmps := make([]app.InstallComponent, 0)
	for _, cmp := range parentApp.Components {
		installCmps = append(installCmps, app.InstallComponent{
			ComponentID: cmp.ID,
		})
	}
	install := app.Install{
		AppID:             appID,
		Name:              req.Name,
		Status:            "queued",
		StatusDescription: "waiting for event loop to start and provision install",
		AWSAccount: app.AWSAccount{
			Region:     req.AWSAccount.Region,
			IAMRoleARN: req.AWSAccount.IAMRoleARN,
		},
		AppSandboxConfigID: parentApp.AppSandboxConfigs[0].ID,
		AppRunnerConfigID:  parentApp.AppRunnerConfigs[0].ID,
		InstallComponents:  installCmps,
	}

	res = s.db.WithContext(ctx).Create(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install: %w", res.Error)
	}
	return &install, nil
}
