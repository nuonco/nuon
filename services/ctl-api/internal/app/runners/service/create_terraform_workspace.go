package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateTerraformWorkspaceRequest struct {
	OwnerID   string `json:"owner_id" binding:"required"`
	OwnerType string `json:"owner_type" binding:"required"`
}

// @ID						CreateTerraformWorkspace
// @Summary				create terraform workspace
// @Description.markdown	create_terraform_workspace.md
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
// @Success				200	{array}	app.TerraformWorkspace
// @Router					/v1/terraform-workspace [post]
func (s *service) CreateTerraformWorkpace(ctx *gin.Context) {
	var workspaceReq *CreateTerraformWorkspaceRequest
	err := ctx.ShouldBindJSON(&workspaceReq)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to bind request: %w", err))
		return
	}

	// TODO(ht): validate owner_id and owner_type
	workspace, err := s.createWorkspace(ctx, workspaceReq)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create workspace: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, workspace)
}

func (s *service) createWorkspace(ctx *gin.Context, workspaceReq *CreateTerraformWorkspaceRequest) (*app.TerraformWorkspace, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err

	}

	workspace := &app.TerraformWorkspace{
		OwnerID:   workspaceReq.OwnerID,
		OwnerType: app.TerraformWorkspaceOwner(workspaceReq.OwnerType),
		OrgID:     orgID,
	}
	err = s.db.WithContext(ctx).Create(workspace).Error
	if err != nil {
		return nil, err
	}

	return workspace, nil
}
