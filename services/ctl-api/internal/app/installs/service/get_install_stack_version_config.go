package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// installStackVersionConfigResponse carries the rendered install-stack config
// block with no run — this is a read-only fetch used by the Terraform
// provider's nuon_stack data source so an install-stacks module can read its
// configuration from the API instead of having it rendered into tfvars.
type installStackVersionConfigResponse struct {
	Config *app.InstallerSDKConfig `json:"config"`
}

// @ID						GetInstallStackVersionConfig
// @Summary				get the SDK config for a stack version
// @Description			return the rendered install-stack configuration (runner, permissions, inputs, secrets) for a stack version. Read-only and side-effect free. Public endpoint: the per-stack-version phone_home_id in the URL is the secret. Consumed by the Terraform provider's nuon_stack data source.
// @Param					phone_home_id	path	string	true	"stack version phone-home ID (used as the URL secret)"
// @Tags					stacks
// @Accept					json
// @Produce				json
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	installStackVersionConfigResponse
// @Router					/v1/stack-runs/{phone_home_id}/config [get]
func (s *service) GetInstallStackVersionConfig(ctx *gin.Context) {
	phoneHomeID := ctx.Param("phone_home_id")

	var stackVersion app.InstallStackVersion
	if res := s.db.WithContext(ctx).
		Where(app.InstallStackVersion{PhoneHomeID: phoneHomeID}).
		First(&stackVersion); res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			ctx.Error(stderr.ErrNotFound{Err: res.Error, Description: "stack version not found"})
			return
		}
		ctx.Error(fmt.Errorf("load stack version: %w", res.Error))
		return
	}

	// Public endpoint: org context isn't set by middleware, but the config
	// builder's helper chain needs it. Set it from the loaded stack version.
	reqCtx := cctx.SetOrgIDContext(ctx.Request.Context(), stackVersion.OrgID)
	reqCtx = cctx.SetAccountIDContext(reqCtx, stackVersion.CreatedByID)

	cfg, err := s.helpers.BuildInstallerSDKConfig(reqCtx, stackVersion.InstallID)
	if err != nil {
		ctx.Error(fmt.Errorf("build installer sdk config: %w", err))
		return
	}
	cfg.PhoneHomeURL = stackVersion.PhoneHomeURL

	s.helpers.ApplyInstallInputValues(reqCtx, cfg, stackVersion.InstallID)

	ctx.JSON(http.StatusOK, installStackVersionConfigResponse{Config: cfg})
}
