package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
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

type gcpGARImageConfigRequest struct {
	GCPProjectID             string `json:"gcp_project_id"`
	GCPRegion                string `json:"gcp_region"`
	ImageURL                 string `json:"image_url"`
	Tag                      string `json:"tag"`
	ServiceAccountEmail      string `json:"service_account_email,omitempty"`
	WorkloadIdentityProvider string `json:"workload_identity_provider,omitempty"`
}

func (g *gcpGARImageConfigRequest) getGCPGARImageConfig() *app.GCPGARImageConfig {
	if g == nil {
		return nil
	}

	return &app.GCPGARImageConfig{
		GCPProjectID:             g.GCPProjectID,
		GCPRegion:                g.GCPRegion,
		ServiceAccountEmail:      g.ServiceAccountEmail,
		WorkloadIdentityProvider: g.WorkloadIdentityProvider,
	}
}

type azureACRImageConfigRequest struct {
	RegistryURL string `json:"registry_url"`
	TenantID    string `json:"tenant_id,omitempty"`
	ClientID    string `json:"client_id,omitempty"`
}

func (a *azureACRImageConfigRequest) getAzureACRImageConfig() *app.AzureACRImageConfig {
	if a == nil {
		return nil
	}

	return &app.AzureACRImageConfig{
		RegistryURL: a.RegistryURL,
		TenantID:    a.TenantID,
		ClientID:    a.ClientID,
	}
}

type CreateExternalImageComponentConfigRequest struct {
	AWSECRImageConfig   *awsECRImageConfigRequest   `json:"aws_ecr_image_config"`
	GCPGARImageConfig   *gcpGARImageConfigRequest   `json:"gcp_gar_image_config"`
	AzureACRImageConfig *azureACRImageConfigRequest `json:"azure_acr_image_config"`

	ImageURL      string `json:"image_url" validate:"required"`
	Tag           string `json:"tag" validate:"required"`
	BuildTimeout  string `json:"build_timeout,omitempty"`  // Duration string for build operations (e.g., "30m", "1h")
	DeployTimeout string `json:"deploy_timeout,omitempty"` // Duration string for deploy operations (e.g., "30m", "1h")

	AppConfigID string `json:"app_config_id"`

	Dependencies   []string                      `json:"dependencies"`
	References     []string                      `json:"references"`
	Checksum       string                        `json:"checksum"`
	OperationRoles map[app.OperationType]*string `json:"operation_roles,omitempty"`
}

func (c *CreateExternalImageComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	if c.OperationRoles != nil {
		for operation := range c.OperationRoles {
			if !slices.Contains(app.ValidOperations, operation) {
				return fmt.Errorf("invalid operation type: %s. Valid operations: %v", operation, app.ValidOperations)
			}
		}
	}

	if c.BuildTimeout != "" {
		if err := validateBuildTimeout(c.BuildTimeout); err != nil {
			return err
		}
	}
	if c.DeployTimeout != "" {
		if err := validateDeployTimeout(c.DeployTimeout); err != nil {
			return err
		}
	}
	return nil
}

// @ID						CreateAppExternalImageComponentConfig
// @Summary				create an external image component config
// @Description.markdown	create_external_image_component_config.md
// @Param					req				body	CreateExternalImageComponentConfigRequest	true	"Input"
// @Param					app_id			path	string										true	"app ID"
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
// @Router					/v1/apps/{app_id}/components/{component_id}/configs/external-image [POST]
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
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

	cfg := app.ExternalImageComponentConfig{
		ImageURL:            req.ImageURL,
		Tag:                 req.Tag,
		AWSECRImageConfig:   req.AWSECRImageConfig.getAWSECRImageConfig(),
		GCPGARImageConfig:   req.GCPGARImageConfig.getGCPGARImageConfig(),
		AzureACRImageConfig: req.AzureACRImageConfig.getAzureACRImageConfig(),
	}

	operationRoles := make(pgtype.Hstore)
	for operation, role := range req.OperationRoles {
		operationRoles[string(operation)] = role
	}

	componentConfigConnection := app.ComponentConfigConnection{
		ExternalImageComponentConfig: &cfg,
		ComponentID:                  parentCmp.ID,
		AppConfigID:                  req.AppConfigID,
		ComponentDependencyIDs:       pq.StringArray(depIDs),
		References:                   pq.StringArray(req.References),
		Checksum:                     req.Checksum,
		BuildTimeout:                 req.BuildTimeout,
		DeployTimeout:                req.DeployTimeout,
		OperationRoles:               operationRoles,
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create external image component config connection: %w", res.Error)
	}

	return &cfg, nil
}
