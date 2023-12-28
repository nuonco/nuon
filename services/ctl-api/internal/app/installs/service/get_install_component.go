package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetInstallComponent
// @Summary	get an install component
// @Description.markdown	get_install_component.md
// @Param			install_id		path	string	true	"install ID"
// @Param			component_id	path	string	true	"component ID"
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
// @Success		200				{object}	app.InstallComponent
// @Router			/v1/installs/{install_id}/component/{component_id} [get]
func (s *service) GetInstallComponent(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	installCmp, err := s.getInstallComponent(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get  install cmp %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, installCmp)
}

func (s *service) getInstallComponent(ctx context.Context, installID, componentID string) (*app.Install, error) {
	return nil, nil
}
