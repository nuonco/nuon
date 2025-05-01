package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetTerraformWorkspaces
// @Summary				get  terraform workspaces
// @Description.markdown	get_terraform_workspaces.md
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
// @Success				200	{array}	app.TerraformWorkspace
// @Router					/v1/terraform-workspaces [get]
func (s *service) GetTerraformWorkpaces(ctx *gin.Context) {
	workspaces, err := s.listWorkspaces(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list workspaces: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, workspaces)
}

func (s *service) listWorkspaces(ctx *gin.Context) ([]app.TerraformWorkspace, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var workspaces []app.TerraformWorkspace
	err = s.db.WithContext(ctx).Model(&app.TerraformWorkspace{}).Where("org_id = ?", orgID).Find(&workspaces).Error
	if err != nil {
		return nil, err
	}

	return workspaces, nil
}
