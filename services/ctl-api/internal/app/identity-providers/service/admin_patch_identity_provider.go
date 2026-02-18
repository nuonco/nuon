package service

import (
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

// AdminPatchIdentityProviderRequest represents the request to update an identity provider.
type AdminPatchIdentityProviderRequest struct {
	Enabled *bool `json:"enabled,omitempty"`

	// Provider-specific config fields (only one should be set based on existing provider_type)
	OpenIDConfig *providers.OpenIDConfig `json:"openid_config,omitempty"`
	GoogleConfig *providers.GoogleConfig `json:"google_config,omitempty"`
	GitHubConfig *providers.GitHubConfig `json:"github_config,omitempty"`
}

func (r *AdminPatchIdentityProviderRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	return nil
}

// @ID						AdminPatchIdentityProvider
// @Summary				Update an existing identity provider
// @Description.markdown	admin_patch_identity_provider.md
// @Param					identity_provider_id	path	string								true	"identity provider ID"
// @Param					req						body	AdminPatchIdentityProviderRequest	true	"Input"
// @Tags					auth/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.IdentityProvider
// @Router					/v1/auth/identity-providers/{identity_provider_id} [PATCH]
func (s *service) AdminPatchIdentityProvider(ctx *gin.Context) {
	identityProviderID := ctx.Param("identity_provider_id")
	if identityProviderID == "" {
		ctx.Error(fmt.Errorf("identity_provider_id is required"))
		return
	}

	var req AdminPatchIdentityProviderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Load existing provider
	var ip app.IdentityProvider
	result := s.db.WithContext(ctx).Where("id = ?", identityProviderID).First(&ip)
	if result.Error != nil {
		ctx.Error(fmt.Errorf("identity provider not found: %w", result.Error))
		return
	}

	// Update enabled if provided
	if req.Enabled != nil {
		ip.Enabled = *req.Enabled
	}

	// Update config if provided (based on existing provider type)
	var configErr error
	switch ip.ProviderType {
	case app.ProviderTypeOIDC:
		if req.OpenIDConfig != nil {
			configErr = ip.SetOpenIDConfig(req.OpenIDConfig)
		}
	case app.ProviderTypeGoogle:
		if req.GoogleConfig != nil {
			configErr = ip.SetGoogleConfig(req.GoogleConfig)
		}
	case app.ProviderTypeGitHub:
		if req.GitHubConfig != nil {
			configErr = ip.SetGitHubConfig(req.GitHubConfig)
		}
	}
	if configErr != nil {
		ctx.Error(fmt.Errorf("failed to set provider config: %w", configErr))
		return
	}

	// Validate the config if it was updated
	if req.OpenIDConfig != nil || req.GoogleConfig != nil || req.GitHubConfig != nil {
		if err := ip.ValidateConfig(); err != nil {
			ctx.Error(fmt.Errorf("invalid provider config: %w", err))
			return
		}
	}

	// Save the updated provider
	if err := s.db.WithContext(ctx).Save(&ip).Error; err != nil {
		s.l.Error("failed to update identity provider",
			zap.String("id", identityProviderID),
			zap.Error(err))
		ctx.Error(fmt.Errorf("failed to update identity provider: %w", err))
		return
	}

	s.l.Info("updated identity provider",
		zap.String("id", ip.ID),
		zap.String("provider_type", string(ip.ProviderType)),
		zap.Bool("enabled", ip.Enabled))

	ctx.JSON(http.StatusOK, ip)
}
