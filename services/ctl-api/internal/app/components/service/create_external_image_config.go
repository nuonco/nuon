package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
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
	basicVCSConfigRequest
	BasicDeployConfig *basicDeployConfigRequest `json:"basic_deploy_config" validate:"required_unless=SyncOnly true"`

	AWSECRImageConfig *awsECRImageConfigRequest `json:"aws_ecr_image_config"`

	SyncOnly bool   `json:"sync_only"`
	ImageURL string `json:"image_url" validate:"required"`
	Tag      string `json:"tag" validate:"required"`
}

func (c *CreateExternalImageComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @BasePath /v1/components

// Create an external image component config
// @Summary create an external image component config
// @Schemes
// @Description create an external image component config.
// @Param req body CreateExternalImageComponentConfigRequest true "Input"
// @Param component_id path string component_id "component ID"
// @Tags components
// @Accept json
// @Produce json
// @Success 201 {object} app.ComponentConfigConnection
// @Router /v1/components/{component_id}/configs/external-image [POST]
func (s *service) CreateExternalImageComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")
	if cmpID == "" {
		ctx.Error(fmt.Errorf("component id must be passed in"))
		return
	}

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

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) createExternalImageComponentConfig(ctx context.Context, cmpID string, req *CreateExternalImageComponentConfigRequest) (*app.ExternalImageComponentConfig, error) {
	parentCmp, err := s.getComponentWithParents(ctx, cmpID)
	if err != nil {
		return nil, err
	}

	// build component config
	connectedGithubVCSConfig, err := req.connectedGithubVCSConfig(parentCmp)
	if err != nil {
		return nil, fmt.Errorf("invalid connected github vcs config: %w", err)
	}
	cfg := app.ExternalImageComponentConfig{
		PublicGitVCSConfig:       req.publicGitVCSConfig(),
		ConnectedGithubVCSConfig: connectedGithubVCSConfig,

		ImageURL:          req.ImageURL,
		Tag:               req.Tag,
		SyncOnly:          req.SyncOnly,
		BasicDeployConfig: req.BasicDeployConfig.getBasicDeployConfig(),
		AWSECRImageConfig: req.AWSECRImageConfig.getAWSECRImageConfig(),
	}

	componentConfigConnection := app.ComponentConfigConnection{
		ExternalImageComponentConfig: &cfg,
		ComponentID:                  parentCmp.ID,
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create external image component config connection: %w", res.Error)
	}

	return &cfg, nil
}
