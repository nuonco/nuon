package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

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

// @ID UpdateInstallInputs
// @Summary	Updates install input config for app
// @Description.markdown	update_install_inputs.md
// @Tags			installs
// @Accept			json
// @Param			req	body	UpdateInstallInputsRequest	true	"Input"
// @Produce		json
// @Param			install_id		path		string	true	"install ID"
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	app.InstallInputs
// @Router			/v1/installs/{install_id}/inputs [patch]
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

	installInputs, err := s.getInstallInputs(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install inputs: %w", err))
		return
	}

	// if no inputs, exit early
	if len(installInputs) < 1 {
		ctx.Error(fmt.Errorf("no inputs found for install: %w", gorm.ErrRecordNotFound))
		return
	}

	latestInstallInput := installInputs[0]

	err = s.validateInstallInput(ctx, latestInstallInput.AppInputConfigID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to validate install input: %w", err))
		return
	}

	inputs, err := s.newInstallInputs(ctx, latestInstallInput, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install inputs: %w", err))
		return
	}

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type: signals.OperationDeployComponents,
	})

	ctx.JSON(http.StatusOK, inputs)
}

func (s *service) validateInstallInput(ctx context.Context, appInputConfigID string, req UpdateInstallInputsRequest) error {
	appInputs := []*app.AppInput{}
	res := s.db.WithContext(ctx).
		Find(&appInputs, "app_input_config_id = ?", appInputConfigID)
	if res.Error != nil {
		return fmt.Errorf("unable to get app inputs: %w", res.Error)
	}

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

func (s *service) newInstallInputs(ctx context.Context, installInput app.InstallInputs, req UpdateInstallInputsRequest) (*app.InstallInputs, error) {
	inputs := map[string]*string{}
	for k, v := range installInput.Values {
		inputs[k] = v
	}

	for k, v := range req.Inputs {
		inputs[k] = v
	}

	// this update will be tied to the same AppInputConfigID tied to the latest install input
	obj := &app.InstallInputs{
		AppInputConfigID: installInput.AppInputConfigID,
		InstallID:        installInput.InstallID,
		Values:           pgtype.Hstore(inputs),
	}
	res := s.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install inputs: %w", res.Error)
	}

	return obj, nil
}
