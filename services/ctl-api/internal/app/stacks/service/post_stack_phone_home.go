package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/stackrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// StackPhoneHomeRequest is the body of a stack phone home: the stack's outputs plus
// a `request_type` naming the lifecycle event. Same shape as the legacy route's body.
type StackPhoneHomeRequest = installshelpers.StackPhoneHomeRequest

// @ID						PostStackPhoneHome
// @Summary				phone home for an install stack
// @Description			Report an install stack's outputs for the install's latest stack version. Authenticated replacement for the public capability-URL route: the caller's token identifies the stack's service account, so no phone_home_id appears in the path.
// @Param					install_id	path	string				true	"install ID"
// @Param					req			body	StackPhoneHomeRequest	true	"Input"
// @Tags					stacks/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.EmptyResponse
// @Router					/v1/stacks/{install_id}/phone-home [post]
func (s *service) PostStackPhoneHome(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to resolve org from request: %w", err))
		return
	}

	var req StackPhoneHomeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	requestType, ok := req["request_type"].(string)
	if !ok {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("request type param must be a string")))
		return
	}
	if !installshelpers.ValidPhoneHomeRequestType(requestType) {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("invalid request type %q", requestType)))
		return
	}

	// Same org-scoped lookup as GetStackConfig: not-found rather than forbidden, so
	// this cannot probe install IDs in other orgs.
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

	// Delete carries nothing the control plane records — matching the legacy route,
	// which accepts and drops it so a deprovisioned install's stack stays deletable.
	if requestType == installshelpers.PhoneHomeRequestTypeDelete {
		ctx.JSON(http.StatusOK, app.EmptyResponse{})
		return
	}

	// Latest version, the one being applied. Unlike the config read — where a missing
	// version is normal, because the module reads config before the version exists —
	// a report has nowhere to land without one, so this is a not-found the module retries.
	var version app.InstallStackVersion
	if res := s.db.WithContext(ctx).
		Where(app.InstallStackVersion{InstallID: install.ID}).
		Order("created_at DESC").
		First(&version); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{Err: res.Error, Description: "install stack version not found"})
			return
		}
		ctx.Error(fmt.Errorf("load install stack version: %w", res.Error))
		return
	}

	run, err := s.installsHelpers.RecordStackPhoneHome(ctx.Request.Context(), &version, req)
	if err != nil {
		ctx.Error(fmt.Errorf("record stack phone home: %w", err))
		return
	}

	reqCtx := cctx.SetOrgIDContext(ctx.Request.Context(), version.OrgID)
	reqCtx = cctx.SetAccountIDContext(reqCtx, version.CreatedByID)
	if err := s.installsHelpers.EnqueueInstallSignal(reqCtx, install.ID,
		installshelpers.InstallSignalsQueueName, &stackrun.Signal{
			InstallStackID:        version.InstallStackID,
			InstallStackVersionID: version.ID,
			RunID:                 run.ID,
			RequestType:           requestType,
		}); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, app.EmptyResponse{})
}
