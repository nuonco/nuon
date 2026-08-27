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
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

// StackPhoneHomeRequest is the body of a stack phone home: the stack's outputs plus
// a `request_type` naming the lifecycle event. Same shape as the legacy route's body.
type StackPhoneHomeRequest = installshelpers.StackPhoneHomeRequest

// stackPhoneHomeInputsKey carries the input values the stack resolved. A report of
// inputs, not a stack output, so it is stripped from the payload.
const stackPhoneHomeInputsKey = "inputs"

// extractStackPhoneHomeInputs pulls the optional `inputs` object off the body and
// removes it. Absent is fine; a malformed one is a user error.
func extractStackPhoneHomeInputs(req StackPhoneHomeRequest) (map[string]string, error) {
	raw, ok := req[stackPhoneHomeInputsKey]
	if !ok {
		return nil, nil
	}
	delete(req, stackPhoneHomeInputsKey)

	if raw == nil {
		return nil, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("inputs must be an object of strings")
	}

	inputs := make(map[string]string, len(obj))
	for name, v := range obj {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("input %q must be a string", name)
		}
		inputs[name] = str
	}
	return inputs, nil
}

// @ID						PostStackPhoneHome
// @Summary				phone home for an install stack
// @Description			Report an install stack's outputs for the install's latest stack version. Authenticated replacement for the public capability-URL route: the caller's token identifies the stack's service account, so no phone_home_id appears in the path. An optional `inputs` object of string values reports the install-input values the stack resolved; every key must be a customer-source app input, and the merged result becomes the install's current inputs.
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

	inputs, err := extractStackPhoneHomeInputs(req)
	if err != nil {
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

	// Accepted and dropped, like the legacy route, so a deprovisioned stack stays
	// deletable.
	if requestType == installshelpers.PhoneHomeRequestTypeDelete {
		ctx.JSON(http.StatusOK, app.EmptyResponse{})
		return
	}

	// Latest version, the one being applied. A report has nowhere to land without
	// one, so unlike the config read this is a not-found.
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

	// Before the run: a rejected report must not leave a run claiming values never
	// persisted. The create hooks need org and account, which the request context drops.
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to resolve account from request: %w", err))
		return
	}
	inputsCtx := cctx.SetOrgIDContext(ctx.Request.Context(), orgID)
	inputsCtx = cctx.SetAccountIDContext(inputsCtx, acct.ID)
	_, inputWorkflow, err := s.installsHelpers.SetInstallInputsFromStack(inputsCtx, &install, inputs)
	if err != nil {
		ctx.Error(err)
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

	// The workflow was created above; this makes it run.
	if inputWorkflow != nil {
		if err := s.installsHelpers.EnqueueInstallSignal(reqCtx, install.ID,
			installshelpers.InstallWorkflowsQueueName, &executeflow.Signal{
				WorkflowID: inputWorkflow.ID,
			}); err != nil {
			ctx.Error(fmt.Errorf("enqueue input update workflow signal: %w", err))
			return
		}
	}

	ctx.JSON(http.StatusCreated, app.EmptyResponse{})
}
