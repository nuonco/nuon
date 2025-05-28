package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID						DeleteTerraformStateJSON
// @Summary				delete terraform state json
// @Description.markdown	delete_terraform_state_json.md
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
// @Router					/v1/terraform-workspaces/{workspace_id}/state [delete]
func (s *service) DeleteTerraformWorkspaceStateJSON(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})
		return
	}
	err := s.helpers.DeleteStateJSON(ctx, workspaceID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to delete state json: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "terraform state json deleted"})
}
