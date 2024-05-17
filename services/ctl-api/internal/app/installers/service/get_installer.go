package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID GetInstaller
// @Summary	get an installer
// @Description.markdown	get_installer.md
// @Param			installer_id	path	string	true	"installer ID"
// @Tags installers
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	app.Installer
// @Router			/v1/installers/{installer_id} [get]
func (s *service) GetInstaller(ctx *gin.Context) {
	installerID := ctx.Param("installer_id")
	app, err := s.getInstaller(ctx, installerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installer %s: %w", installerID, err))
		return
	}

	ctx.JSON(http.StatusOK, app)
}
