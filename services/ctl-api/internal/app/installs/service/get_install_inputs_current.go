package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @ID GetCurrentInstallInputs
// @Summary	get an installs current inputs
// @Description.markdown	get_install_inputs.md
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Param			install_id		path		string	true	"install ID"
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	app.InstallInputs
// @Router			/v1/installs/{install_id}/inputs/current [GET]
func (s *service) GetInstallCurrentInputs(ctx *gin.Context) {
	appID := ctx.Param("install_id")

	installInputs, err := s.getInstallInputs(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install inputs: %w", err))
		return
	}

	if len(installInputs) < 1 {
		ctx.Error(fmt.Errorf("no inputs found for install: %w", gorm.ErrRecordNotFound))
		return
	}

	ctx.JSON(http.StatusOK, installInputs[0])
}
