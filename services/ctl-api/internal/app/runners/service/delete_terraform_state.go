package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID						DeleteTerraformState
// @Summary				delete terraform state
// @Description.markdown	delete_terraform_state.md
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
// @Success				200	{object}	string
// @Router					/v1/terraform-backend [delete]
func (s *service) DeleteTerraformState(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "")
}
