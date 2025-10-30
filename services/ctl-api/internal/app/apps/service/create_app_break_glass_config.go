package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateAppBreakGlassConfigRequest struct {
	Roles []AppAWSIAMRoleConfig `json:"roles" validate:"required"`

	AppConfigID string `json:"app_config_id" validate:"required"`
}

func (c *CreateAppBreakGlassConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	return nil
}

// @ID						CreateAppBreakGlassConfig
// @Description.markdown	create_app_break_glass_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppBreakGlassConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppBreakGlassConfig
// @Router /v1/apps/{app_id}/break-glass-configs [post]
func (s *service) CreateAppBreakGlasssConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppBreakGlassConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: err,
		})
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.createAppBreakGlassConfig(ctx, appID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createAppBreakGlassConfig(ctx context.Context, appID string, req *CreateAppBreakGlassConfigRequest) (*app.AppBreakGlassConfig, error) {
	obj := app.AppBreakGlassConfig{
		AppID:       appID,
		AppConfigID: req.AppConfigID,
		Roles:       make([]app.AppAWSIAMRoleConfig, 0),
	}

	for _, role := range req.Roles {
		obj.Roles = append(obj.Roles, app.AppAWSIAMRoleConfig{
			AppConfigID:             req.AppConfigID,
			Type:                    app.AWSIAMRoleTypeBreakGlass,
			Name:                    role.Name,
			Description:             role.Description,
			DisplayName:             role.DisplayName,
			PermissionsBoundaryJSON: generics.ToJSON(role.PermissionsBoundary),
			Policies:                role.getPolicies(req.AppConfigID),
		})
	}

	res := s.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create app break glass config")
	}

	return &obj, nil
}
