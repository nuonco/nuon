package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
)

type CreateInstallerInstallRequest struct {
	helpers.CreateInstallParams
}

func (c *CreateInstallerInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID InstallerCreateInstall
// @Summary	create an app install from an installer
// @Description.markdown installer_create_install.md
// @Param			req	body	CreateInstallRequest	true	"Input"
// @Tags			installers
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
	var req CreateInstallerInstallRequest
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

	install, err := s.installHelpers.CreateInstall(cctx, installer.App.ID, &req.CreateInstallParams)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	s.installHooks.Created(ctx, install.ID, installer.App.Org.OrgType)

	ctx.JSON(http.StatusCreated, install)
}
