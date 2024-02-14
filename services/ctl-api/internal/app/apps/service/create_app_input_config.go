package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	validatoradapter "github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/validator"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

type AppInputRequest struct {
	DisplayName string `json:"display_name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Default     string `json:"default"`
	Required    bool   `json:"required"`
}

type CreateAppInputConfigRequest struct {
	Inputs map[string]AppInputRequest `json:"inputs" validate:"required"`
}

func (c *CreateAppInputConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	for k := range c.Inputs {
		if err := validatoradapter.InterpolatedName(v, k); err != nil {
			return fmt.Errorf("invalid input %s - %w", k, err)
		}
	}

	return nil
}

// @ID CreateAppInputConfig
// @Description.markdown	create_app_input_config.md
// @Tags			apps
// @Accept			json
// @Param			req	body	CreateAppInputConfigRequest	true	"Input"
// @Produce		json
// @Param			app_id	path	string				true	"app ID"
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.AppInputConfig
// @Router			/v1/apps/{app_id}/input-config [post]
func (s *service) CreateAppInputsConfig(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")

	var req CreateAppInputConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	inputs, err := s.createAppInputs(ctx, org.ID, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app inputs config: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, inputs)
}

func (s *service) createAppInputs(ctx context.Context, orgID, appID string, req *CreateAppInputConfigRequest) (*app.AppInputConfig, error) {
	appInputs := make([]app.AppInput, 0, len(req.Inputs))
	for name, input := range req.Inputs {
		appInputs = append(appInputs, app.AppInput{
			Name:        name,
			Description: input.Description,
			DisplayName: input.DisplayName,
			Required:    input.Required,
			Default:     input.Default,
		})
	}

	inputs := app.AppInputConfig{
		OrgID:     orgID,
		AppID:     appID,
		AppInputs: appInputs,
	}

	res := s.db.WithContext(ctx).Create(&inputs)
	if res.Error != nil {

		return nil, fmt.Errorf("unable to create app inputs: %w", res.Error)
	}

	return &inputs, nil
}
