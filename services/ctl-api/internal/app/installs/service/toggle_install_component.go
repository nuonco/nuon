package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type ToggleInstallComponentRequest struct {
	Enabled  bool   `json:"enabled"`
	Role     string `json:"role,omitempty"`
	PlanOnly bool   `json:"plan_only"`
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
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
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

	installConfig, err := s.helpers.GetLatestInstallConfig(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install config: %w", err))
		return
	}

	currentlyEnabled := true
	if installConfig != nil {
		currentlyEnabled = installConfig.IsComponentEnabled(componentID, latestConfig)
	} else {
		currentlyEnabled = latestConfig.GetDefaultEnabled()
	}

	if currentlyEnabled == req.Enabled {
		state := "enabled"
		if !req.Enabled {
			state = "disabled"
		}
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("component already %s", state),
			Description: fmt.Sprintf("This component is already %s", state),
		})
		return
	}

	if err := s.updateComponentToggle(ctx, installID, componentID, req.Enabled, installConfig); err != nil {
		ctx.Error(fmt.Errorf("unable to update component toggle: %w", err))
		return
	}

	workflowType := app.WorkflowTypeComponentEnabled
	if !req.Enabled {
		workflowType = app.WorkflowTypeComponentDisabled
	}

	workflow, err := s.helpers.CreateWorkflowWithRole(ctx,
		install.ID,
		workflowType,
		map[string]string{
			"component_id": component.ID,
		},
		req.PlanOnly,
		req.Role,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	queueID, err := s.getInstallWorkflowsQueueID(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, &executeflow.Signal{
		WorkflowID: workflow.ID,
	}, workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, app.WorkflowResponse{WorkflowID: workflow.ID})
}

func (s *service) updateComponentToggle(ctx *gin.Context, installID, componentID string, enabled bool, existingConfig *app.InstallConfig) error {
	if existingConfig == nil {
		installConfig := &app.InstallConfig{
			InstallID:        installID,
			ApprovalOption:   app.InstallApprovalOptionPrompt,
			ComponentToggles: map[string]bool{componentID: enabled},
		}
		return s.db.WithContext(ctx).Create(installConfig).Error
	}

	toggles := existingConfig.ComponentToggles
	if toggles == nil {
		toggles = make(map[string]bool)
	}
	toggles[componentID] = enabled

	return s.db.WithContext(ctx).
		Model(existingConfig).
		Update("component_toggles", toggles).
		Error
}
