package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/updated"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateInstallInputsRequest struct {
	Inputs           map[string]*string `json:"inputs" validate:"required,gte=1"`
	Role             string             `json:"role"`
	DeployDependents bool               `json:"deploy_dependents"`
}

func (c *UpdateInstallInputsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						UpdateInstallInputs
// @Summary				Updates install input config for app
// @Description.markdown	update_install_inputs.md
// @Tags					installs
// @Accept					json
// @Param					req	body	UpdateInstallInputsRequest	true	"Input"
// @Produce				json
// @Param					install_id	path	string	true	"install ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallInputs
// @Router					/v1/installs/{install_id}/inputs [patch]
func (s *service) UpdateInstallInputs(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req UpdateInstallInputsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if len(install.App.AppInputConfigs) < 1 {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("no app input configs defined on app"),
			Description: "no app input configs defined",
		})
		return
	}

	latestLatestInstallInputs, err := s.getLatestInstallInputs(ctx, install.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get latest install inputs: %w", err))
		return
	}

	pinnedAppInputConfig, err := s.helpers.GetPinnedAppInputConfig(ctx, install.AppID, install.AppConfigID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get latest app input config: %w", err))
		return
	}

	err = s.helpers.ValidateInstallInputs(ctx, pinnedAppInputConfig, req.Inputs)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Validate that only vendor inputs are being updated
	err = s.validateVendorSourceInputs(pinnedAppInputConfig, req.Inputs)
	if err != nil {
		ctx.Error(err)
		return
	}

	inputs, changedInputs, changedInputValues, err := s.newInstallInputs(
		ctx,
		latestLatestInstallInputs,
		pinnedAppInputConfig,
		req,
	)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install inputs: %w", err))
		return
	}

	workflow, err := s.helpers.CreateAndStartInputUpdateWorkflow(
		ctx,
		install.ID,
		*changedInputs,
		changedInputValues,
		req.Role,
		req.DeployDependents,
	)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install inputs: %w", err))
		return
	}

	// Enqueue queue signals so the input-update workflow runs.
	signalsQueueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	workflowsQueueID, err := s.getInstallWorkflowsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, signalsQueueID, &updated.Signal{
		InstallID: install.ID,
	}, "", ""); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}
	if err := s.enqueueInstallSignal(ctx, workflowsQueueID, &executeflow.Signal{
		WorkflowID: workflow.ID,
	}, workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	inputs.WorkflowID = &workflow.ID
	ctx.JSON(http.StatusOK, inputs)
}

func (s *service) getLatestInstallInputs(ctx context.Context, installID string) (*app.InstallInputs, error) {
	installInputs := app.InstallInputs{}
	res := s.db.WithContext(ctx).
		Where("install_id = ?", installID).
		Order("created_at DESC").
		First(&installInputs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install inputs: %w", res.Error)
	}

	return &installInputs, nil
}

func (s *service) getLatestAppInputConfig(ctx context.Context, appID string) (*app.AppInputConfig, error) {
	appInputConfig := app.AppInputConfig{}
	res := s.db.WithContext(ctx).
		Preload("AppInputs").
		Where("app_id = ?", appID).
		Order("created_at DESC").
		First(&appInputConfig)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app input config: %w", res.Error)
	}

	return &appInputConfig, nil
}

func (s *service) newInstallInputs(
	ctx context.Context,
	installInputs *app.InstallInputs,
	appInputConfig *app.AppInputConfig,
	req UpdateInstallInputsRequest,
) (*app.InstallInputs, *[]string, string, error) {
	inputs := map[string]*string{}
	for k, v := range installInputs.Values {
		inputs[k] = v
	}

	for k, v := range req.Inputs {
		inputs[k] = v
	}

	// create a lookup for the latest app input config
	appInputNames := map[string]struct{}{}
	for _, input := range appInputConfig.AppInputs {
		appInputNames[input.Name] = struct{}{}
	}

	changed, err := helpers.ComputeChangedInputs(
		installInputs.Values,
		req.Inputs,
		appInputConfig.AppInputs,
	)
	if err != nil {
		return nil, nil, "", fmt.Errorf("unable to compute changed inputs: %w", err)
	}

	// remove inputs not in the latest app input config
	for k := range inputs {
		_, ok := appInputNames[k]
		if !ok {
			delete(inputs, k)
			continue
		}
	}

	// this update will be tied to the latest AppInputConfigID for the app
	obj := &app.InstallInputs{
		AppInputConfigID: appInputConfig.ID,
		InstallID:        installInputs.InstallID,
		Values:           pgtype.Hstore(inputs),
	}
	res := s.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return nil, nil, "", fmt.Errorf("unable to create install inputs: %w", res.Error)
	}

	latestInstallInputs, err := s.getLatestInstallInputs(ctx, installInputs.InstallID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("unable to get latest install inputs: %w", err)
	}

	latestInstallInputs.Values = nil

	return latestInstallInputs, &changed.Names, changed.ChangedValuesJSON, nil
}

func (s *service) validateVendorSourceInputs(appInputConfig *app.AppInputConfig, inputs map[string]*string) error {
	appInputSources := map[string]app.AppInputSource{}
	for _, input := range appInputConfig.AppInputs {
		appInputSources[input.Name] = input.Source
	}

	for name := range inputs {
		source, ok := appInputSources[name]
		if !ok {
			return fmt.Errorf("input %s is not defined in app input config", name)
		}

		// Reject customer sourced inputs
		if source == app.AppInputSourceCustomer {
			return fmt.Errorf("%s has source install_stack, cannot be updated via api", name)
		}
	}

	return nil
}
