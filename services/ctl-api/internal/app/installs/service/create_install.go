package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

type CreateInstallRequest struct {
	helpers.CreateInstallParams
}

func (c *CreateInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID CreateInstall
// @Summary	create an app install
// @Description.markdown	create_install.md
// @Param			app_id	path	string					true	"app ID"
// @Param			req		body	CreateInstallRequest	true	"Input"
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
// @Success		201				{object}	app.Install
// @Router			/v1/apps/{app_id}/installs [post]
func (s *service) CreateInstall(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")

	var req CreateInstallRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	install, err := s.helpers.CreateInstall(ctx, appID, &req.CreateInstallParams)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	s.hooks.Created(ctx, install.ID, org.OrgType)
	ctx.JSON(http.StatusCreated, install)
}
