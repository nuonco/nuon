package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateInstallConfigRequest struct {
	helpers.CreateInstallConfigParams
}

func (c *CreateInstallConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateInstallConfig
// @Summary				create an install config
// @Description.markdown	create_install_config.md
// @Tags					installs
// @Param					install_id	path	string	true	"install ID"
// @Accept					json
// @Param					req	body	CreateInstallConfigRequest	true	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.InstallConfig
// @Router					/v1/installs/{install_id}/configs [post]
func (s *service) CreateInstallConfig(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateInstallConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	app, err := s.helpers.CreateInstallConfig(ctx, installID, &req.CreateInstallConfigParams)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, app)
}
