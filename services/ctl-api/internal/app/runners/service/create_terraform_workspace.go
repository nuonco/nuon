package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateTerraformWorkspaceRequest struct {
	OwnerID   string `json:"owner_id" binding:"required"`
	OwnerType string `json:"owner_type" binding:"required"`
}

// @ID						CreateTerraformWorkspace
// @Summary				create terraform workspace
// @Description.markdown	create_terraform_workspace.md
// @Param					req	body	CreateTerraformWorkspaceRequest	true	"Input"
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
// @Success				201	{object}	app.TerraformWorkspace
// @Router 				/v1/terraform-workspace [post]
func (s *service) CreateTerraformWorkspace(ctx *gin.Context) {
	var req CreateTerraformWorkspaceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(err)
		return
	}

	workspace := app.TerraformWorkspace{
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
	}

	if err := s.db.WithContext(ctx).Create(&workspace).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, workspace)
}

// @ID						CreateTerraformWorkspaceV2
// @Summary				create terraform workspace
// @Description.markdown	create_terraform_workspace.md
// @Param					req	body	CreateTerraformWorkspaceRequest	true	"Input"
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
// @Success				201	{object}	app.TerraformWorkspace
// @Router 				/v1/terraform-workspaces [post]
func (s *service) CreateTerraformWorkspaceV2(ctx *gin.Context) {
	var req CreateTerraformWorkspaceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(err)
		return
	}

	workspace := app.TerraformWorkspace{
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
	}

	if err := s.db.WithContext(ctx).Create(&workspace).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, workspace)
}
