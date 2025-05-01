package service

import (
	"encoding/json"
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
// @Tags					runners,runners/runner
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
// @Success				200	{object}	app.TerraformWorkspaceState
// @Router					/v1/terraform-backend [post]
func (s *service) UpdateTerraformState(ctx *gin.Context) {
	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})
		return
	}

	reqLockID := ctx.Query("ID")
	if reqLockID != "" {
		currLock, err := s.helpers.GetWorkspaceLock(ctx, reqLockID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get lock: %w", err))
			return
		}

		if currLock != nil && currLock.ID != reqLockID {
			ctx.Error(stderr.ErrInvalidRequest{
				Err: fmt.Errorf("lock ID does not match current lock: %s", reqLockID),
			})
			return

		}
	}

	// Get the raw body first
	contents, err := ctx.GetRawData()
	if err != nil {
		ctx.Error(fmt.Errorf("unable to read request body: %w", err))
		return
	}
	var data app.TerraformStateData

	if err := json.Unmarshal(contents, &data); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	_, err = s.helpers.GetTerraformState(ctx, workspaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Error(err)
		return
	}

	_, err = s.helpers.InsertTerraformState(ctx, workspaceID, contents, &data)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update terraform state: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "")
}
