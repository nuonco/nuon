package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminUnlockWorkspace struct{}

// @ID						AdminUnlockWorkspace
// @Summary				unlock a terraform workspace
// @Description.markdown admin_unlock_workspace.md
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req			body	AdminDeleteRunnerRequest	true	"Input"
// @Param workspace_id	path	string	true	"workspace ID or owner ID of workspace to unlock"
// @Produce				json
// @Success				200	{string}	ok
// @Router					/v1/terraform-workspaces/{workspace_id}/unlock [post]
func (s *service) AdminUnlockWorkspace(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")

	workspace, err := s.findWorkspace(ctx, workspaceID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.helpers.UnlockWorkspace(ctx, workspace.ID); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}
