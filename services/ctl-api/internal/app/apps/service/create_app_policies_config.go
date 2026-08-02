package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateAppPoliciesConfigRequest struct {
	AppConfigID string `json:"app_config_id" validate:"required"`

	Policies []AppPolicyConfig `json:"policies"`
}

func (c CreateAppPoliciesConfigRequest) buildInputs() []build.PolicyInput {
	objs := make([]build.PolicyInput, 0, len(c.Policies))
	for _, policy := range c.Policies {
		objs = append(objs, build.PolicyInput{
			Type:        policy.Type,
			Engine:      policy.Engine,
			Name:        policy.Name,
			Description: policy.Description,
			Contents:    policy.Contents,
			Components:  policy.Components,
		})
	}
	return objs
}

type AppPolicyConfig struct {
	Type        config.AppPolicyType   `json:"type" validate:"required"`
	Engine      config.AppPolicyEngine `json:"engine,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Contents    string                 `json:"contents" validate:"required" swaggertype:"string"`
	Components  []string               `json:"components,omitempty"`
}

func (c *CreateAppPoliciesConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateAppPoliciesConfig
// @Description.markdown	create_app_policies_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppPoliciesConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppPoliciesConfig
// @Router /v1/apps/{app_id}/policies-configs [post]
func (s *service) CreateAppPoliciesConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppPoliciesConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: err,
		})
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.createAppPoliciesConfig(ctx, appID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createAppPoliciesConfig(ctx context.Context, appID string, req *CreateAppPoliciesConfigRequest) (*app.AppPoliciesConfig, error) {
	obj, err := build.PoliciesConfig(req.buildInputs(), appID, req.AppConfigID)
	if err != nil {
		return nil, stderr.ErrUser{Err: err, Description: err.Error()}
	}

	res := s.db.WithContext(ctx).Create(obj)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create app policies config")
	}

	return obj, nil
}
