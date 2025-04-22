package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminaBackfillInstallComponentsRequest struct{}

type AdminaBackfillInstallComponentsResponse struct {
	InstallComponentIDs []string `json:"install_component_ids"`
}

// @ID						AdminaBackfillInstallComnponents
// @Description.markdown	admin_backfill_install_components.md
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminaBackfillInstallComponentsRequest true	"Input"
// @Produce				json
// @Success				200	{object}	AdminaBackfillInstallComponentsResponse
// @Router					/v1/installs/admin-backfill-install-components [POST]
func (s *service) AdminaBackfillInstallComponents(ctx *gin.Context) {
	// get installs without install_sandbox_runs
	ics := []app.InstallComponent{}
	res := s.db.WithContext(ctx).
		Where("status = ?", app.InstallComponentStatusUnset).
		Limit(100).
		Find(&ics)
	if res.Error != nil {
		ctx.Error(errors.Wrap(res.Error, "unable to get installs"))
		return
	}

	for _, ic := range ics {
		deploys := []app.InstallDeploy{}
		res = s.db.WithContext(ctx).
			Where("install_component_id = ?", ic.ID).
			Order("created_at desc").
			Limit(1).
			Find(&deploys)

		if res.Error != nil {
			ctx.Error(errors.Wrap(res.Error, "unable to get deploy"))
			return
		}

		status := app.InstallComponentStatusUnknown
		if len(deploys) > 0 {
			status = app.DeployStatusToComponentStatus(deploys[0].Status)
		}

		res = s.db.WithContext(ctx).
			Unscoped().
			Model(&app.InstallComponent{}).
			Where("id = ?", ic.ID).
			UpdateColumn("status", status).
			UpdateColumn("status_description", deploys[0].StatusDescription)

		if res.Error != nil {
			ctx.Error(errors.Wrap(res.Error, "unable to create install component"))
			return
		}
	}

	installComponentIDs := []string{}
	for _, ic := range ics {
		installComponentIDs = append(installComponentIDs, ic.ID)
	}

	ctx.JSON(http.StatusOK, AdminaBackfillInstallComponentsResponse{
		InstallComponentIDs: installComponentIDs,
	})
}
