package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppInstallInputConfig
// @Summary				get the input config a new install would be created with
// @Description.markdown	get_app_install_input_config.md
// @Param					app_id			path	string	true	"app ID"
// @Param					app_branch_id	query	string	false	"app branch ID"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.AppInputConfig
// @Router					/v1/apps/{app_id}/install-input-config [get]
func (s *service) GetAppInstallInputConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	appBranchID := ctx.Query("app_branch_id")

	if _, err := s.findApp(ctx, org.ID, appID); err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	inputConfig, err := s.installsHelpers.CreateInstallInputConfig(ctx, org.ID, appID, appBranchID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, inputConfig)
}
