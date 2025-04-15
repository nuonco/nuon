package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

type UpdateInstallInputsRequest struct {
	Inputs map[string]*string `json:"inputs" validate:"required,gte=1"`
}

func (c *UpdateInstallInputsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
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

	latestLatestInstallInputs, err := s.getLatestInstallInputs(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get latest install inputs: %w", err))
		return
	}

	latestAppInputConfig, err := s.getLatestAppInputConfig(ctx, install.AppID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get latest app input config: %w", err))
		return
	}

	err = s.validateInstallInput(ctx, *latestAppInputConfig, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to validate install input: %w", err))
		return
	}

	inputs, err := s.newInstallInputs(ctx, *latestLatestInstallInputs, *latestAppInputConfig, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install inputs: %w", err))
		return
	}

	// TODO(jm): remove this once the legacy install flow is deprecated
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureIndependentRunner)
	if err != nil {
		ctx.Error(err)
		return
	}
	if !enabled {
		ctx.JSON(http.StatusOK, inputs)
		return
	}

	workflow, err := s.helpers.CreateInstallWorkflow(ctx, install.ID, app.InstallWorkflowTypeInputUpdate, map[string]string{
		// NOTE(jm): this metadata field is not really designed to be used for anything serious, outside of
		// rendering things in the UI and other such things, which is why we are just using a string slice here,
		// maybe that will change at some point, but this metadata should not be abused.
		"inputs": strings.Join(generics.MapToKeys(req.Inputs), ","),
	}, app.StepErrorBehaviorAbort)
	if err != nil {
		ctx.Error(err)
		return
	}
	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type:              signals.OperationExecuteWorkflow,
		InstallWorkflowID: workflow.ID,
	})

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

func (s *service) validateInstallInput(ctx context.Context, appInputConfig app.AppInputConfig, req UpdateInstallInputsRequest) error {
	appInputs := appInputConfig.AppInputs
	appInputNames := map[string]struct{}{}
	for _, input := range appInputs {
		appInputNames[input.Name] = struct{}{}
	}

	for name := range req.Inputs {
		if _, ok := appInputNames[name]; !ok {
			return fmt.Errorf("name %s does not exist in app inputs", name)
		}
	}

	return nil
}

func (s *service) newInstallInputs(ctx context.Context, installInputs app.InstallInputs, appInputConfig app.AppInputConfig, req UpdateInstallInputsRequest) (*app.InstallInputs, error) {
	inputs := map[string]*string{}
	for k, v := range installInputs.Values {
		inputs[k] = v
	}

	for k, v := range req.Inputs {
		inputs[k] = v
	}

	// create a lookup for the latest app input config
	appInputs := appInputConfig.AppInputs
	appInputNames := map[string]struct{}{}
	for _, input := range appInputs {
		appInputNames[input.Name] = struct{}{}
	}

	// remove inputs not in the latest app input config
	for k := range inputs {
		if _, ok := appInputNames[k]; !ok {
			delete(inputs, k)
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
		return nil, fmt.Errorf("unable to create install inputs: %w", res.Error)
	}

	return obj, nil
}
