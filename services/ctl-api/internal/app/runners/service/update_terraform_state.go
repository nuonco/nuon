package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
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
	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})
		return
	}

	lockID := ctx.Query("ID")
	if lockID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("lock_id was not set"),
		})
		return
	}

	var data app.TerraformStateData
	if err := ctx.BindJSON(&data); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	currentState, err := s.helpers.GetTerraformState(ctx, workspaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Error(err)
		return
	}

	if currentState == nil {
		currentState = &app.TerraformState{}
		currentState.Lock.ID = lockID
	}
	currentState.Data = &data

	_, err = s.helpers.InsertTerraformState(ctx, workspaceID, currentState.Lock, &data)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update terraform state: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "")
}
