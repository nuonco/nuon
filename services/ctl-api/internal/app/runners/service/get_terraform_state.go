package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// @ID						GetTerraformCurrentStateData
// @Summary				get current terraform
// @Description.markdown	get_terraform_current_state.md
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
// @Success				200	{object}	app.TerraformWorkspaceState
// @Router					/v1/terraform-backend [get]
func (s *service) GetTerraformCurrentStateData(ctx *gin.Context) {
	workspaceID := ctx.Query("workspace_id")

	state, err := s.helpers.GetTerraformState(ctx, workspaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Error(err)
		return
	}

	if state == nil || state.Data == nil || state.Data.Version < 1 {
		ctx.JSON(http.StatusNoContent, "")
		return
	}

	ctx.JSON(http.StatusOK, state.Data)
}
