package service

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID						GetTerraformStateJSON
// @Summary				get terraform state json
// @Description.markdown	get_terraform_state_json.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	interface{}
// @Router					/v1/terraform-workspaces/{workspace_id}/state [get]
func (s *service) GetTerraformWorkspaceStateJSON(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})
		return
	}

	state, err := s.helpers.GetTerraformStateJSON(ctx, workspaceID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if state != nil {
		ctx.JSON(http.StatusOK, state)
		return
	}

	ctx.JSON(http.StatusNotFound, gin.H{"error": "terraform state not found"})
}
