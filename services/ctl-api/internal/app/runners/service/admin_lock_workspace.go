package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminLockWorkspace struct{}

// @ID						AdminLockWorkspace
// @Summary				lock a terraform workspace
// @Description.markdown admin_lock_workspace.md
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req			body	AdminDeleteRunnerRequest	true	"Input"
// @Param workspace_id	path	string	true	"workspace ID or owner ID of workspace to unlock"
// @Produce				json
// @Success				200	{string}	ok
// @Router					/v1/terraform-workspaces/{workspace_id}/lock [post]
func (s *service) AdminLockWorkspace(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")

	workspace, err := s.findWorkspace(ctx, workspaceID)
	if err != nil {
		ctx.Error(err)
		return
	}

	lock, err := s.helpers.LockWorkspace(ctx, workspace.ID, nil, &app.TerraformLock{
		Created:   generics.GetFakeObj[string](),
		Path:      generics.GetFakeObj[string](),
		ID:        workspaceID,
		Operation: generics.GetFakeObj[string](),
		Info:      generics.GetFakeObj[string](),
		Who:       generics.GetFakeObj[string](),
		Version:   generics.GetFakeObj[*string](),
	})
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to lock workspace"))
		return
	}

	ctx.JSON(http.StatusOK, lock)
}

func (s *service) findWorkspace(ctx context.Context, workspaceID string) (*app.TerraformWorkspace, error) {
	wkspace := app.TerraformWorkspace{}
	res := s.db.WithContext(ctx).
		Where("id = ?", workspaceID).
		Or("owner_id = ?", workspaceID).
		First(&wkspace)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find workspace")
	}

	return &wkspace, nil
}
