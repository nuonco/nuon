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

	identities, err := s.getAccountIdentities(ctx, account.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	filteredIdentities := make([]AuthMeIdentity, len(identities))
	for i, identity := range identities {
		filteredIdentities[i] = AuthMeIdentity{
			ProviderType: identity.ProviderType,
			Name:         identity.Name,
			Picture:      identity.Picture,
		}
	}

	response := AuthMeResponse{
		Account:    account,
		Identities: filteredIdentities,
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *service) getAccountIdentities(ctx *gin.Context, accountID string) ([]app.AccountIdentity, error) {
	var identities []app.AccountIdentity
	res := s.db.WithContext(ctx).
		Where(&app.AccountIdentity{AccountID: accountID}).
		Find(&identities)
	if res.Error != nil {
		return nil, res.Error
	}
	return identities, nil
}
