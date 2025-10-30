package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateAppPoliciesConfigRequest struct {
	AppConfigID string `json:"app_config_id" validate:"required"`

	Policies []AppPolicyConfig `json:"policies"`
}

func (c CreateAppPoliciesConfigRequest) getPolicies(appID, appConfigID string) []app.AppPolicyConfig {
	objs := make([]app.AppPolicyConfig, 0)
	for _, policy := range c.Policies {
		objs = append(objs, app.AppPolicyConfig{
			AppID:       appID,
			AppConfigID: appConfigID,
			Type:        policy.Type,
			Contents:    policy.Contents,
		})
	}
	return objs
}

type AppPolicyConfig struct {
	Type     config.AppPolicyType `json:"type" validate:"required"`
	Contents string               `json:"contents" validate:"required" swaggertype:"string"`
}

func (c *CreateAppPoliciesConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	for _, policy := range c.Policies {
		if !generics.SliceContains(policy.Type, app.AllPolicyTypes) {
			return stderr.ErrUser{
				Err:         fmt.Errorf("policy type must be one of (%s)", strings.Join(generics.ToStringSlice(app.AllPolicyTypes), ",")),
				Description: "invalid policy type " + string(policy.Type),
			}
		}
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
	obj := app.AppPoliciesConfig{
		AppID:       appID,
		AppConfigID: req.AppConfigID,
		Policies:    req.getPolicies(appID, req.AppConfigID),
	}

	res := s.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create app break glass config")
	}

	return &obj, nil
}
