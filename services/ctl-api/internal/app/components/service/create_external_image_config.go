package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type awsECRImageConfigRequest struct {
	IAMRoleARN string `json:"iam_role_arn"`
	AWSRegion  string `json:"aws_region"`
}

func (a *awsECRImageConfigRequest) getAWSECRImageConfig() *app.AWSECRImageConfig {
	if a == nil {
		return nil
	}

	return &app.AWSECRImageConfig{
		IAMRoleARN: a.IAMRoleARN,
		AWSRegion:  a.AWSRegion,
	}
}

type CreateExternalImageComponentConfigRequest struct {
	AWSECRImageConfig *awsECRImageConfigRequest `json:"aws_ecr_image_config"`

	ImageURL string `json:"image_url" validate:"required"`
	Tag      string `json:"tag" validate:"required"`

	AppConfigID string `json:"app_config_id"`

	Dependencies []string `json:"dependencies"`
	References   []string `json:"references"`
	Checksum     string   `json:"checksum"`
}

func (c *CreateExternalImageComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateAppExternalImageComponentConfig
// @Summary				create an external image component config
// @Description.markdown	create_external_image_component_config.md
// @Param					req				body	CreateExternalImageComponentConfigRequest	true	"Input"
// @Param					component_id	path	string										true	"component ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.ExternalImageComponentConfig
// @Router					/v1/components/{component_id}/configs/external-image [POST]
func (s *service) CreateAppExternalImageComponentConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	cmpID := ctx.Param("component_id")
	_, err := s.getAppComponent(ctx, appID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	// reuse the same logic as non-app scoped endpoint
	s.CreateExternalImageComponentConfig(ctx)
}

// @ID						CreateExternalImageComponentConfig
// @Summary				create an external image component config
// @Description.markdown	create_external_image_component_config.md
// @Param					req				body	CreateExternalImageComponentConfigRequest	true	"Input"
// @Param					component_id	path	string										true	"component ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.ExternalImageComponentConfig
// @Router					/v1/components/{component_id}/configs/external-image [POST]
func (s *service) CreateExternalImageComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	var req CreateExternalImageComponentConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.createExternalImageComponentConfig(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component cfg: %w", err))
		return
	}

	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type: signals.OperationConfigCreated,
	})

	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type:          signals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeExternalImage,
	})

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createExternalImageComponentConfig(ctx context.Context, cmpID string, req *CreateExternalImageComponentConfigRequest) (*app.ExternalImageComponentConfig, error) {
	parentCmp, err := s.getComponentWithParents(ctx, cmpID)
	if err != nil {
		return nil, err
	}

	depIDs, err := s.helpers.GetComponentIDs(ctx, parentCmp.AppID, req.Dependencies)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component ids")
	}

	// build component config
	cfg := app.ExternalImageComponentConfig{
		ImageURL:          req.ImageURL,
		Tag:               req.Tag,
		AWSECRImageConfig: req.AWSECRImageConfig.getAWSECRImageConfig(),
	}

	componentConfigConnection := app.ComponentConfigConnection{
		ExternalImageComponentConfig: &cfg,
		ComponentID:                  parentCmp.ID,
		AppConfigID:                  req.AppConfigID,
		ComponentDependencyIDs:       pq.StringArray(depIDs),
		References:                   pq.StringArray(req.References),
		Checksum:                     req.Checksum,
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create external image component config connection: %w", res.Error)
	}

	return &cfg, nil
}
