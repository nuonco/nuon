package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID				DeleteRunbook
// @Summary		delete a runbook
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			app_id		path	string	true	"app ID"
// @Param			runbook_id	path	string	true	"runbook ID"
// @Success		200			{object}	bool
// @Router			/v1/apps/{app_id}/runbooks/{runbook_id} [delete]
func (s *service) DeleteRunbook(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	runbookID := ctx.Param("runbook_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	res := s.db.WithContext(ctx).
		Where(app.Runbook{OrgID: org.ID}).
		Delete(&app.Runbook{}, "id = ?", runbookID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to delete runbook: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, true)
}
