package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// AdminCreateIdentityProviderRequest represents the request to create an identity provider.
type AdminCreateIdentityProviderRequest struct {
	ProviderType string `json:"provider_type" validate:"required,oneof=oidc google github"`
	Enabled      bool   `json:"enabled"`

	// Name labels the provider on the sign-in page. Several providers can share a provider_type,
	// so this is what tells them apart.
	Name string `json:"name,omitempty"`

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
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.IdentityProvider
// @Router					/v1/auth/identity-providers [POST]
func (s *service) AdminCreateIdentityProvider(ctx *gin.Context) {
	var req AdminCreateIdentityProviderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	providerType := app.ProviderType(req.ProviderType)

	// Build the identity provider
	ip := &app.IdentityProvider{
		ProviderType: providerType,
		Name:         req.Name,
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

	// Several providers can now share a provider_type, so the only thing worth rejecting is the
	// same application registered twice.
	duplicate, err := s.findDuplicateProvider(ctx, ip)
	if err != nil {
		ctx.Error(fmt.Errorf("failed to check for duplicate identity provider: %w", err))
		return
	}
	if duplicate != nil {
		s.l.Warn("duplicate identity provider",
			zap.String("provider_type", req.ProviderType),
			zap.String("existing_id", duplicate.ID))
		ctx.Error(stderr.ErrConflict{
			Err:         fmt.Errorf("an identity provider with this client_id already exists (id: %s)", duplicate.ID),
			Description: "identity provider already exists",
		})
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

// findDuplicateProvider reports an existing global provider registered against the same
// application: same type, same client ID and, for OIDC, same issuer.
func (s *service) findDuplicateProvider(ctx context.Context, ip *app.IdentityProvider) (*app.IdentityProvider, error) {
	clientID, err := ip.GetClientID()
	if err != nil {
		return nil, err
	}

	var issuerURL string
	if ip.ProviderType == app.ProviderTypeOIDC {
		cfg, err := ip.GetOpenIDConfig()
		if err != nil {
			return nil, err
		}
		issuerURL = cfg.IssuerURL
	}

	if string(ip.ProviderType) == s.cfg.NuonAuthProviderType &&
		clientID == s.cfg.NuonAuthClientID &&
		(ip.ProviderType != app.ProviderTypeOIDC || issuerURL == s.cfg.NuonAuthIssuerURL) {
		return &app.IdentityProvider{
			ID:           app.EnvIdentityProviderID(ip.ProviderType),
			ProviderType: ip.ProviderType,
		}, nil
	}

	var existing []app.IdentityProvider
	if err := s.db.WithContext(ctx).
		Where(&app.IdentityProvider{ProviderType: ip.ProviderType}).
		Where("org_id IS NULL").
		Find(&existing).Error; err != nil {
		return nil, err
	}

	for i := range existing {
		existingClientID, err := existing[i].GetClientID()
		if err != nil || existingClientID != clientID {
			continue
		}

		if ip.ProviderType != app.ProviderTypeOIDC {
			return &existing[i], nil
		}

		cfg, err := existing[i].GetOpenIDConfig()
		if err == nil && cfg.IssuerURL == issuerURL {
			return &existing[i], nil
		}
	}

	return nil, nil
}
