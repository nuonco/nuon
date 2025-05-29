package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

// @ID						GetTerraformWorkspaceLock
// @Summary				get terraform workspace lock
// @Description.markdown	get_terraform_workspace_lock.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.TerraformWorkspaceLock
// @Router					/v1/terraform-workspaces/{workspace_id}/lock [get]
func (s *service) GetTerraformWorkspaceLock(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})
		return
	}

	tfl, err := s.getTerraformWorkspaceLock(ctx, workspaceID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to lock workspace: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, tfl)
}

func (s *service) getTerraformWorkspaceLock(ctx *gin.Context, workspaceID string) (*app.TerraformWorkspaceLock, error) {
	tfs := &app.TerraformWorkspaceLock{}

	res := s.db.WithContext(ctx).
		Scopes(runnerJobPreload).
		First(tfs, "workspace_id = ?", workspaceID)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}

	return tfs, nil
}
