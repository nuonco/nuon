package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// Keeps the removed /v1/stack-runs/{phone_home_id}/config response shape, which the
// stack SDK's decoder was built against.
type stackConfigResponse struct {
	Config *app.InstallerSDKConfig `json:"config"`
}

// @ID						GetStackConfig
// @Summary				get the SDK config for an install stack
// @Description			Return the rendered install-stack configuration (runner, permissions, inputs, secrets) for an install, including the phone-home URL the stack reports completion to. Read-only and side-effect free. Authenticated: the caller's token identifies the stack's service account, or an OIDC-federated account with access to the org.
// @Param					install_id	path	string	true	"install ID"
// @Tags					stacks/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	stackConfigResponse
// @Router					/v1/stacks/{install_id}/config [get]
func (s *service) GetStackConfig(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to resolve org from request: %w", err))
		return
	}

	// Org scope is defense in depth; not-found so this cannot probe other orgs.
	var install app.Install
	if res := s.db.WithContext(ctx).
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&install); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{Err: res.Error, Description: "install not found"})
			return
		}
		ctx.Error(fmt.Errorf("load install: %w", res.Error))
		return
	}

	// BuildInstallerSDKConfig reads org and account from the context, and gin does not
	// carry them into ctx.Request.Context().
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to resolve account from request: %w", err))
		return
	}
	reqCtx := cctx.SetOrgIDContext(ctx.Request.Context(), orgID)
	reqCtx = cctx.SetAccountIDContext(reqCtx, acct.ID)

	cfg, err := s.installsHelpers.BuildInstallerSDKConfig(reqCtx, install.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("build installer sdk config: %w", err))
		return
	}

	// The authenticated report route. CloudFormation and ARM keep the capability URL.
	cfg.PhoneHomeURL = fmt.Sprintf("%s/v1/stacks/%s/phone-home",
		strings.TrimSuffix(cfg.RunnerAPIURL, "/"), install.ID)

	ctx.JSON(http.StatusOK, stackConfigResponse{Config: cfg})
}
