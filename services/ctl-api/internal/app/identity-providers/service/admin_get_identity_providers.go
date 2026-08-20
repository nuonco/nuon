package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	identityProviderSourceEnv      = "env"
	identityProviderSourceDatabase = "database"
)

// AdminIdentityProviderSummary is a slim representation of an identity provider
// exposing only the fields needed to determine whether a provider exists or is enabled.
type AdminIdentityProviderSummary struct {
	ID            string           `json:"id"`
	ProviderType  app.ProviderType `json:"provider_type"`
	Name          string           `json:"name,omitempty"`
	ClientID      string           `json:"client_id"`
	Enabled       bool             `json:"enabled"`
	AllowAllUsers *bool            `json:"allow_all_users,omitempty"`

	// Source is "env" for the provider configured through environment variables, which has no
	// row in identity_providers, and "database" for the rest.
	Source string `json:"source"`
}

// @ID						AdminGetIdentityProviders
// @Summary				List identity providers
// @Description.markdown	admin_get_identity_providers.md
// @Tags					auth/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	AdminIdentityProviderSummary
// @Router					/v1/auth/identity-providers [GET]
func (s *service) AdminGetIdentityProviders(ctx *gin.Context) {
	var ips []app.IdentityProvider
	if err := s.db.WithContext(ctx).Find(&ips).Error; err != nil {
		s.l.Error("failed to list identity providers", zap.Error(err))
		ctx.Error(fmt.Errorf("failed to list identity providers: %w", err))
		return
	}

	summaries := make([]AdminIdentityProviderSummary, 0, len(ips)+1)
	if envSummary := s.envIdentityProviderSummary(); envSummary != nil {
		summaries = append(summaries, *envSummary)
	}

	for i := range ips {
		clientID, err := ips[i].GetClientID()
		if err != nil {
			s.l.Warn("failed to parse client_id for identity provider",
				zap.String("id", ips[i].ID),
				zap.Error(err))
		}

		summaries = append(summaries, AdminIdentityProviderSummary{
			ID:            ips[i].ID,
			ProviderType:  ips[i].ProviderType,
			Name:          ips[i].Name,
			ClientID:      clientID,
			Enabled:       ips[i].Enabled,
			AllowAllUsers: ips[i].AllowAllUsers,
			Source:        identityProviderSourceDatabase,
		})
	}

	ctx.JSON(http.StatusOK, summaries)
}

// envIdentityProviderSummary reports the provider configured through environment variables. It has
// no row in identity_providers, so without this the endpoint an operator uses to check what is
// configured omits the provider that is always present.
func (s *service) envIdentityProviderSummary() *AdminIdentityProviderSummary {
	providerType := app.ProviderType(s.cfg.NuonAuthProviderType)
	if providerType == "" {
		return nil
	}

	return &AdminIdentityProviderSummary{
		ID:           app.EnvIdentityProviderID(providerType),
		ProviderType: providerType,
		Name:         s.cfg.NuonAuthProviderName,
		ClientID:     s.cfg.NuonAuthClientID,
		Enabled:      true,
		Source:       identityProviderSourceEnv,
	}
}
