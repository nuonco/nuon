package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

// @ID						DeleteInstallSandbox
// @Summary				delete an install component
// @Description.markdown	delete_install_sandbox.md
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
// @Router					/v1/installs/{install_id}/sandbox [delete]
func (s *service) DeleteInstallSandbox(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	force := ctx.DefaultQuery("force", "false") == "true"

	_, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type:        signals.OperationDeprovision,
		ForceDelete: force,
	})

	ctx.JSON(http.StatusOK, true)
}
