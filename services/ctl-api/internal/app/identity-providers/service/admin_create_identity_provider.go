package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// AdminCreateIdentityProviderRequest represents the request to create an identity provider.
type AdminCreateIdentityProviderRequest struct {
	ProviderType string `json:"provider_type" validate:"required,oneof=oidc google github"`
	Enabled      bool   `json:"enabled"`

	// Provider-specific config fields (only one should be set based on provider_type)
	OpenIDConfig *providers.OpenIDConfig `json:"openid_config,omitempty"`
	GoogleConfig *providers.GoogleConfig `json:"google_config,omitempty"`
	GitHubConfig *providers.GitHubConfig `json:"github_config,omitempty"`
}

func (r *AdminCreateIdentityProviderRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	// Ensure the correct config is provided for the provider type
	switch r.ProviderType {
	case "oidc":
		if r.OpenIDConfig == nil {
			return fmt.Errorf("openid_config is required for oidc provider type")
		}
	case "google":
		if r.GoogleConfig == nil {
			return fmt.Errorf("google_config is required for google provider type")
		}
	case "github":
		if r.GitHubConfig == nil {
			return fmt.Errorf("github_config is required for github provider type")
		}
	}

	return nil
}

// @ID						AdminCreateIdentityProvider
// @Summary				Create a new identity provider
// @Description.markdown	admin_create_identity_provider.md
// @Param					req	body	AdminCreateIdentityProviderRequest	true	"Input"
// @Tags					auth/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.IdentityProvider
// @Router					/v1/auth/identity-providers [POST]
func (s *service) AdminCreateIdentityProvider(ctx *gin.Context) {
	var req AdminCreateIdentityProviderRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	providerType := app.ProviderType(req.ProviderType)

	// Check if this provider type conflicts with the default env var provider
	if s.cfg.NuonAuthProviderType == req.ProviderType {
		s.l.Warn("attempt to create provider with same type as default",
			zap.String("provider_type", req.ProviderType))
		ctx.Error(fmt.Errorf("provider type %s is already configured as the default provider via environment variables", req.ProviderType))
		return
	}

	// Check if a provider of this type already exists in the database
	var existing app.IdentityProvider
	result := s.db.WithContext(ctx).
		Where("provider_type = ? AND org_id IS NULL", providerType).
		First(&existing)
	if result.Error == nil {
		s.l.Warn("provider type already exists",
			zap.String("provider_type", req.ProviderType),
			zap.String("existing_id", existing.ID))
		ctx.Error(fmt.Errorf("a global identity provider of type %s already exists (id: %s)", req.ProviderType, existing.ID))
		return
	}

	// Build the identity provider
	ip := &app.IdentityProvider{
		ProviderType: providerType,
		Enabled:      req.Enabled,
		// OrgID is intentionally left empty for global providers
	}

	// Set the config based on provider type and validate it
	var configErr error
	switch providerType {
	case app.ProviderTypeOIDC:
		configErr = ip.SetOpenIDConfig(req.OpenIDConfig)
	case app.ProviderTypeGoogle:
		configErr = ip.SetGoogleConfig(req.GoogleConfig)
	case app.ProviderTypeGitHub:
		configErr = ip.SetGitHubConfig(req.GitHubConfig)
	}
	if configErr != nil {
		ctx.Error(fmt.Errorf("failed to set provider config: %w", configErr))
		return
	}

	// Validate the config using the model's validation method
	if err := ip.ValidateConfig(); err != nil {
		ctx.Error(fmt.Errorf("invalid provider config: %w", err))
		return
	}

	// Create the provider in the database
	if err := s.db.WithContext(ctx).Create(ip).Error; err != nil {
		s.l.Error("failed to create identity provider",
			zap.String("provider_type", req.ProviderType),
			zap.Error(err))
		ctx.Error(fmt.Errorf("failed to create identity provider: %w", err))
		return
	}

	s.l.Info("created identity provider",
		zap.String("id", ip.ID),
		zap.String("provider_type", req.ProviderType),
		zap.Bool("enabled", ip.Enabled))

	ctx.JSON(http.StatusCreated, ip)
}
