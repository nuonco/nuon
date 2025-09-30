package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/app-branches/signals"
)

type AdminRestartAppBranchRequest struct{}

// @ID						AdminRestartAppBranch
// @Summary				restart an app branch event loop
// @Description.markdown	admin_restart_app_branch.md
// @Param					app_branch_id	path	string				true	"app branch ID"
// @Param					req		body	AdminRestartAppBranchRequest	false	"Input"
// @Tags					apps/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/app-branches/{app_branch_id}/admin-restart [POST]
func (s *service) AdminRestartAppBranch(ctx *gin.Context) {
	abID := ctx.Param("app_branch_id")
	ab, err := s.getAppBranch(ctx, abID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	s.evClient.Send(ctx, ab.ID, &signals.Signal{
		Type: signals.OperationRestart,
	})
	ctx.JSON(http.StatusOK, true)
}

func (s *service) getAppBranch(ctx context.Context, abID string) (*app.AppBranch, error) {
	ab := app.AppBranch{}
	res := s.db.WithContext(ctx).
		Where("id = ?", abID).
		First(&ab)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app branch: %w", res.Error)
	}

	return &ab, nil
}
