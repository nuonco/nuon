package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/appconfiggraph"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type ToggleInstallComponentRequest struct {
	Enabled            *bool  `json:"enabled" validate:"required"`
	PlanOnly           bool   `json:"plan_only"`
	TearDownDependents bool   `json:"tear_down_dependents"`
	Role               string `json:"role,omitempty"`
}

func (c *ToggleInstallComponentRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						ToggleInstallComponent
// @Summary				toggle an install component on or off
// @Description			Enable or disable a toggleable component on an install. Enabling triggers a deploy workflow, disabling triggers a teardown workflow.
// @Param					install_id		path	string							true	"install ID"
// @Param					component_id	path	string							true	"component ID"
// @Param					req				body	ToggleInstallComponentRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.WorkflowResponse
// @Router					/v1/installs/{install_id}/components/{component_id}/toggle [post]
func (s *service) ToggleInstallComponent(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	var req ToggleInstallComponentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	install, err := s.helpers.GetInstall(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	component, err := s.helpers.GetComponent(ctx, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	var latestConfig *app.ComponentConfigConnection
	if len(component.ComponentConfigs) > 0 {
		latestConfig = &component.ComponentConfigs[0]
	}
	if latestConfig == nil || !latestConfig.IsToggleable() {
		ctx.Error(stderr.ErrUser{
			Err:         errors.New("component is not toggleable"),
			Description: "This component is not configured as toggleable",
		})
		return
	}

	// Enabled-state is the synthetic enabled install input. Writing it through
	// the normal install-input update flow lets the input-update workflow
	// reconcile the deploy/teardown (and enable/disable lifecycle) for us.
	patch := map[string]*string{
		config.EnabledOverrideInputName(component.Name): generics.ToPtr(strconv.FormatBool(*req.Enabled)),
	}
	tearDownDependents := !*req.Enabled && req.TearDownDependents
	if tearDownDependents {
		dependents, err := s.disabledDependentToggles(ctx, install, component.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		for name, value := range dependents {
			patch[name] = value
		}
	}

	// Drive the toggle through the shared install-inputs update flow, but tag
	// the workflow with a dedicated type so it surfaces as "Enabling/Disabling
	// component" in the UI rather than a generic input update. The synthetic
	// enabled input remains the source of truth; the dedicated workflows
	// delegate to the same reconcile logic.
	workflowType := app.WorkflowTypeComponentEnabled
	deployDependents := true
	if !*req.Enabled {
		workflowType = app.WorkflowTypeComponentDisabled
		deployDependents = false
	}

	inputs, err := s.applyInstallInputsUpdate(ctx, install, patch, req.Role, deployDependents, tearDownDependents, false, req.PlanOnly, workflowType, nil)
	if err != nil {
		ctx.Error(err)
		return
	}

	workflowID := ""
	if inputs.WorkflowID != nil {
		workflowID = *inputs.WorkflowID
	}
	ctx.JSON(http.StatusCreated, app.WorkflowResponse{WorkflowID: workflowID})
}

func (s *service) disabledDependentToggles(ctx context.Context, install *app.Install, componentID string) (map[string]*string, error) {
	appCfg, err := s.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, true)
	if err != nil {
		return nil, fmt.Errorf("unable to get app config: %w", err)
	}
	cccByID := make(map[string]*app.ComponentConfigConnection, len(appCfg.ComponentConfigConnections))
	for i := range appCfg.ComponentConfigConnections {
		ccc := &appCfg.ComponentConfigConnections[i]
		cccByID[ccc.ComponentID] = ccc
	}
	patch := map[string]*string{}
	for _, id := range appconfiggraph.NewGraph(cccByID).TransitiveDependentsClosure([]string{componentID}) {
		if id == componentID {
			continue
		}
		ccc := cccByID[id]
		if ccc == nil || !ccc.IsToggleable() {
			continue
		}
		patch[config.EnabledOverrideInputName(ccc.Component.Name)] = generics.ToPtr("false")
	}
	return patch, nil
}
