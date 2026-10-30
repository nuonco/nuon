package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateRequest struct {
	Name                     string                    `json:"name"`
	Platform                 app.CloudPlatform         `json:"platform" swaggertype:"string" enums:"aws,gcp,azure"`
	TargetID                 string                    `json:"target_id"`
	Principal                string                    `json:"principal"`
	DefaultRegion            string                    `json:"default_region,omitempty"`
	Preset                   app.CloudConnectionPreset `json:"preset"`
	ECRRepositoryARNs        []string                  `json:"ecr_repository_arns,omitempty"`
	GCPProjectNumber         string                    `json:"gcp_project_number,omitempty"`
	GARRepositories          []string                  `json:"gar_repositories,omitempty"`
	AzureTenantID            string                    `json:"azure_tenant_id,omitempty"`
	AzureRegistryID          string                    `json:"azure_registry_id,omitempty"`
	AzureRegistryLoginServer string                    `json:"azure_registry_login_server,omitempty"`
	ACRRepositories          []string                  `json:"acr_repositories,omitempty"`
}

// @ID CreateCloudConnection
// @Summary create a cloud connection
// @Description Create an org-bound AWS, GCP, or Azure identity. Optional repository arrays generate initial read-only setup material but are not stored or used as an allowlist. ECR and GAR use default_region for image connection matching. ACR requires a subscription UUID, existing Entra application client ID and service principal, tenant UUID, registry ARM ID, and bare login server. Its registry must already use RBAC Registry + ABAC Repository Permissions; setup never changes registry mode. Cloud IAM authorizes image pulls. Custom renders AWS trust only.
// @Tags cloud-connections
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param req body CreateRequest true "Input"
// @Failure 400 {object} stderr.ErrResponse
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Success 201 {object} ConnectionResponse
// @Router /v1/cloud-connections [post]
func (s *service) Create(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	connection := app.CloudConnection{
		OrgID: org.ID, Name: req.Name, Platform: req.Platform, TargetID: req.TargetID, Principal: req.Principal,
		DefaultRegion: req.DefaultRegion, Preset: req.Preset,
		GCPProjectNumber: req.GCPProjectNumber,
		AzureTenantID:    req.AzureTenantID, AzureRegistryID: req.AzureRegistryID,
		AzureRegistryLoginServer: req.AzureRegistryLoginServer,
	}
	var repositories []string
	for _, scope := range []struct {
		preset       app.CloudConnectionPreset
		repositories []string
	}{
		{app.CloudConnectionPresetECR, req.ECRRepositoryARNs},
		{app.CloudConnectionPresetGAR, req.GARRepositories},
		{app.CloudConnectionPresetACR, req.ACRRepositories},
	} {
		if len(scope.repositories) == 0 {
			continue
		}
		if scope.preset != connection.Preset {
			ctx.Error(userError(fmt.Errorf("repository setup input must match the connection's %s preset", connection.Preset)))
			return
		}
		repositories = scope.repositories
	}
	if connection.DefaultRegion == "" && len(repositories) > 0 {
		switch connection.Preset {
		case app.CloudConnectionPresetECR:
			repository, _ := arn.Parse(repositories[0])
			connection.DefaultRegion = repository.Region
		case app.CloudConnectionPresetGAR:
			parts := strings.Split(repositories[0], "/")
			if len(parts) == 6 {
				connection.DefaultRegion = parts[3]
			}
		}
	}
	if err := validateConnection(&connection); err != nil {
		ctx.Error(userError(err))
		return
	}
	for _, repository := range repositories {
		if err := cloudconnections.ValidateImageRepository(&connection, repository); err != nil {
			ctx.Error(userError(err))
			return
		}
	}
	if err := s.db.WithContext(ctx).Create(&connection).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			ctx.Error(stderr.ErrConflict{
				Err:         err,
				Description: "A cloud connection for this principal already exists in this organization. Use the existing connection or choose a different principal.",
			})
			return
		}
		ctx.Error(fmt.Errorf("unable to create cloud connection: %w", err))
		return
	}
	if _, err := s.helpers.EnsureConnectionQueue(ctx, &connection); err != nil {
		ctx.Error(err)
		return
	}
	response, err := s.response(ctx, &connection, repositories...)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, response)
}
