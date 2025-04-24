package service

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID						GetTerraformWorkspaceStateResources
// @Summary				get terraform state resources
// @Description.markdown	get_terraform_state_resources.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					state_id 		path	string	true	"state ID"
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	app.TerraformStateResource
// @Router					/v1/runners/terraform-workspace/{workspace_id}/states/{state_id}/resources [get]
func (s *service) GetTerraformWorkspaceStateResources(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")
	stateID := ctx.Param("state_id")
	if workspaceID == "" || stateID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id  or state_id was not set"),
		})

		return
	}

	state, err := s.helpers.GetTerraformStateByID(ctx, workspaceID, ctx.Param("state_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	if state == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "terraform state not found"})
		return
	}

	if state != nil {
		ctx.JSON(http.StatusOK, state.Data.Resources)
		return
	}

	ctx.JSON(http.StatusNotFound, gin.H{"error": "terraform state not found"})
}
