package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// AuthMeIdentity represents a filtered identity for the /auth/me response.
type AuthMeIdentity struct {
	ProviderType app.ProviderType `json:"provider_type"`
	Name         string           `json:"name,omitempty"`
	Picture      string           `json:"picture,omitempty"`
}

// AuthMeResponse is the response for the /v1/auth/me endpoint.
type AuthMeResponse struct {
	*app.Account
	Identities []AuthMeIdentity `json:"identities"`
}

// @ID						GetAuthMe
// @Summary				Get current account with identity information
// @Description			Returns the authenticated account with identity profile information (provider_type, name, picture)
// @Tags					auth
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Success				200	{object}	AuthMeResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/auth/me [GET]
func (s *service) GetAuthMe(ctx *gin.Context) {
	account, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Get full account with identities
	fullAccount, identities, err := s.getAccountWithIdentities(ctx, account.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Filter identities to only include provider_type, name, picture
	filteredIdentities := make([]AuthMeIdentity, len(identities))
	for i, identity := range identities {
		filteredIdentities[i] = AuthMeIdentity{
			ProviderType: identity.ProviderType,
			Name:         identity.Name,
			Picture:      identity.Picture,
		}
	}

	response := AuthMeResponse{
		Account:    fullAccount,
		Identities: filteredIdentities,
	}

	ctx.JSON(http.StatusOK, response)
}

// getAccountWithIdentities fetches an account with its identities.
func (s *service) getAccountWithIdentities(ctx *gin.Context, accountID string) (*app.Account, []app.AccountIdentity, error) {
	var account app.Account

	res := s.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Policies").
		Preload("Roles.Org").
		Preload("Identities").
		Where("id = ?", accountID).
		First(&account)

	if res.Error != nil {
		return nil, nil, res.Error
	}

	identities := account.Identities
	account.Identities = nil // Clear from account since we return separately

	return &account, identities, nil
}
