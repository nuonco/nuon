package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// stackConfigResponse carries the rendered install-stack config. The shape matches
// the older /v1/stack-runs/{phone_home_id}/config response so the stack SDK can move
// across without a decoder change.
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

	// Scoped by org rather than by a token-to-stack binding: the stack's service
	// account is an org admin, so a token that can read this install can already read
	// the org through the public API. A narrower check here would be theatre.
	//
	// Reported as not-found rather than forbidden, so a caller cannot use this
	// endpoint to discover which install IDs exist in other orgs.
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

	// BuildInstallerSDKConfig's helper chain reads org and account from the context,
	// and gin's context does not carry them into ctx.Request.Context(). Both come from
	// the authenticated caller here, rather than being synthesized from the loaded row
	// the way the older public endpoint has to.
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

	// The phone-home URL embeds the stack version's phone_home_id. Serving it here is
	// what lets the module stop taking that ID as an input: it arrives over an
	// authenticated channel instead of being rendered into the customer's tfvars.
	//
	// Latest version, because that is the one the customer is applying. A stack with
	// no version yet is not an error — the module reads its config before the version
	// exists on a first apply, and phones home once it does.
	var version app.InstallStackVersion
	if res := s.db.WithContext(ctx).
		Where(app.InstallStackVersion{InstallID: install.ID}).
		Order("created_at DESC").
		First(&version); res.Error == nil {
		cfg.PhoneHomeURL = version.PhoneHomeURL
	} else if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		ctx.Error(fmt.Errorf("load install stack version: %w", res.Error))
		return
	}

	s.installsHelpers.ApplyInstallInputValues(reqCtx, cfg, install.ID)

	ctx.JSON(http.StatusOK, stackConfigResponse{Config: cfg})
}
