package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

// @ID						DeleteInstallComponents
// @Summary				delete an install component
// @Description.markdown	delete_install_components.md
// @Param					install_id		path	string				true	"install ID"
// @Param					force					query	bool					false	"force delete"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean} 		true
// @Router					/v1/installs/{install_id}/components [delete]
func (s *service) DeleteInstallComponents(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	_, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type: signals.OperationDeleteComponents,
	})

	ctx.JSON(http.StatusOK, true)
}
