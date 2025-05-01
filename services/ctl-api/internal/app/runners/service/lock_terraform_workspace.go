package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID						UpdateTerraformState
// @Summary				update terraform state
// @Description.markdown	lock_terraform_workspace.md
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					body body interface{} true "terraform workspace lock "
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.TerraformWorkspaceState
// @Router					/v1/terraform-workspaces/:workspace_id/lock [post]

func (s *service) LockTerraformWorkspace(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})
		return
	}

	var lock app.TerraformLock
	if err := ctx.BindJSON(&lock); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	_, err := s.helpers.LockWorkspace(ctx, workspaceID, &lock)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to lock workspace: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "")
}
