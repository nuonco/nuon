package service

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						UpdateTerraformState
// @Summary				update terraform state
// @Description.markdown	update_terraform_state.md
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					body body interface{} true "Terraform state data"
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.TerraformState
// @Router					/v1/terraform-backend [post]
func (s *service) UpdateTerraformState(ctx *gin.Context) {
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

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to read request body: %w", err))
		return
	}
	if currentState == nil {
		currentState = &app.TerraformState{}
	}

	currentState.Data = body
	err = s.helpers.InsertTerraformState(ctx, sid, currentState)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update terraform state: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "")
}
