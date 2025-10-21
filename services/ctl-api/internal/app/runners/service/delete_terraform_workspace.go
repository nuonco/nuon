package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						DeleteTerraformWorkspace
// @Summary					delete terraform workspace
// @Description.markdown	delete_terraform_workspace.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Tags						runners,runners/runner
// @Accept						json
// @Produce 					json
// @Security					APIKey
// @Security					OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{array}	app.TerraformWorkspace
// @Router						/v1/terraform-workspaces/{workspace_id} [delete]
func (s *service) DeleteTerraformWorkpace(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")

	workspace, err := s.deleteWorkspace(ctx, workspaceID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get workspace: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, workspace)
}

func (s *service) deleteWorkspace(ctx *gin.Context, id string) (*app.TerraformWorkspace, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var workspace *app.TerraformWorkspace
	err = s.db.WithContext(ctx).Model(&app.TerraformWorkspace{}).Where("id = ? AND org_id = ?", id, orgID).Delete(&workspace).Error
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, fmt.Errorf("workspace not found")
	}
	return workspace, nil
}
