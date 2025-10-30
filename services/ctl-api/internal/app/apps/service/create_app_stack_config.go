package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateAppStackConfigRequest struct {
	Type        app.StackType `json:"type" validate:"required"`
	Description string        `json:"description" validate:"required"`
	Name        string        `json:"name" validate:"required"`

	RunnerNestedTemplateURL string `json:"runner_nested_template_url"`
	VPCNestedTemplateURL    string `json:"vpc_nested_template_url"`

	AppConfigID string `json:"app_config_id" validate:"required"`
}

func (c *CreateAppStackConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateAppStackConfig
// @Summary				create an app stack config
// @Description.markdown	create_app_stack_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppStackConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppStackConfig
// @Router					/v1/apps/{app_id}/stack-configs [post]
func (s *service) CreateAppStackConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppStackConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	runnerConfig, err := s.createAppStackConfig(ctx, appID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, runnerConfig)
}

func (s *service) createAppStackConfig(ctx context.Context, appID string, req *CreateAppStackConfigRequest) (*app.AppStackConfig, error) {
	appCloudFormationStackConfig := app.AppStackConfig{
		Type:                    req.Type,
		AppConfigID:             req.AppConfigID,
		AppID:                   appID,
		Name:                    req.Name,
		Description:             req.Description,
		VPCNestedTemplateURL:    req.VPCNestedTemplateURL,
		RunnerNestedTemplateURL: req.RunnerNestedTemplateURL,
	}
	res := s.db.WithContext(ctx).
		Create(&appCloudFormationStackConfig)
	if res.Error != nil {
		return nil, res.Error
	}

	return &appCloudFormationStackConfig, nil
}
