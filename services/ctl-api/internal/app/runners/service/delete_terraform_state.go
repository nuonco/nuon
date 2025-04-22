package service

import (
	"fmt"
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
	sid, err := s.GetStateID(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get state ID: %w", err))
		return
	}
	lockID := ctx.Query("ID")
	currentState, err := s.validateTerraformStateLock(ctx, sid, lockID)
	if err != nil {
		ctx.Error(err)
		return
	}

	currentState.Data = nil
	err = s.helpers.InsertTerraformState(ctx, sid, currentState)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update terraform state: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, "")
}
