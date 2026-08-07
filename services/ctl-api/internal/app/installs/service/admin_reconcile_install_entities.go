package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminReconcileInstallEntitiesResponse struct {
	InstallID  string `json:"install_id"`
	Components int64  `json:"components"`
	Actions    int64  `json:"actions"`
	Runbooks   int64  `json:"runbooks"`
}

// @ID						AdminReconcileInstallEntities
// @Summary				reconcile an install's components, actions and runbooks
// @Description			Re-derives the install's components, actions and runbooks from its pinned app config. Used to repair installs whose entities were never created.
// @Tags					installs
// @Accept					json
// @Produce				json
// @Param					install_id	path		string	true	"install ID"
// @Success				200			{object}	AdminReconcileInstallEntitiesResponse
// @Failure				400			{object}	stderr.ErrResponse
// @Failure				500			{object}	stderr.ErrResponse
// @Router					/v1/installs/{install_id}/admin-reconcile-entities [post]
func (s *service) AdminReconcileInstallEntities(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	if err := s.helpers.ReconcileInstallComponents(ctx, installID); err != nil {
		ctx.Error(fmt.Errorf("unable to reconcile install components: %w", err))
		return
	}
	if err := s.helpers.ReconcileInstallActions(ctx, installID); err != nil {
		ctx.Error(fmt.Errorf("unable to reconcile install actions: %w", err))
		return
	}
	if err := s.helpers.ReconcileInstallRunbooks(ctx, installID); err != nil {
		ctx.Error(fmt.Errorf("unable to reconcile install runbooks: %w", err))
		return
	}

	resp := AdminReconcileInstallEntitiesResponse{InstallID: installID}
	s.db.WithContext(ctx).Table("install_components").Where("install_id = ?", installID).Count(&resp.Components)
	s.db.WithContext(ctx).Table("install_action_workflows").Where("install_id = ?", installID).Count(&resp.Actions)
	s.db.WithContext(ctx).Table("install_runbooks").Where("install_id = ?", installID).Count(&resp.Runbooks)

	ctx.JSON(http.StatusOK, resp)
}
