package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID						DeleteInstallComponent
// @Summary				delete an install component
// @Description.markdown	delete_install_component.md
// @Param					install_id		path	string				true	"install ID"
// @Param					component_id	path	string				true	"component ID"
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
// @Success				200	{bool} 		true
// @Router					/v1/installs/{install_id}/components/{component_id} [delete]
func (s *service) DeleteInstallComponent(ctx *gin.Context) {
	//installID := ctx.Param("install_id")
	//componentID := ctx.Param("component_id")

	// s.evClient.Send(ctx, installID, &signals.Signal{
	// 	Type:     signals.OperationDeploy,
	// 	DeployID: deploy.ID,
	// })
	ctx.JSON(http.StatusOK, true)
}
