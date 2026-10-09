package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oauthclients"
)

type CreateServiceAccountOAuthClientRequest struct {
	Name string `json:"name"`
}

type CreateOAuthClientSecretRequest struct {
	ExpiresAt *time.Time `json:"expires_at"`
}

type CreateOAuthClientSecretResponse struct {
	Secret       app.OAuthClientSecret `json:"secret"`
	ClientID     string                `json:"client_id"`
	ClientSecret string                `json:"client_secret"`
}

func oauthClientError(err error) error {
	switch {
	case errors.Is(err, oauthclients.ErrClientNotFound), errors.Is(err, oauthclients.ErrSecretNotFound):
		return stderr.ErrNotFound{Err: err, Description: err.Error()}
	case errors.Is(err, oauthclients.ErrTooManySecrets), errors.Is(err, oauthclients.ErrSecretInThePast), errors.Is(err, oauthclients.ErrNotConfidential):
		return stderr.ErrUser{Err: err, Description: err.Error()}
	default:
		return err
	}
}

func (s *service) orgServiceAccountForOAuth(ctx *gin.Context) (*app.Account, string, error) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		return nil, "", err
	}

	caller, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return nil, "", err
	}
	ctx.Request = ctx.Request.WithContext(cctx.SetAccountContext(ctx.Request.Context(), caller))
	acct, err := s.getOrgServiceAccount(ctx.Request.Context(), org.ID, ctx.Param("account_id"))
	if err != nil {
		return nil, "", err
	}

	return acct, org.ID, nil
}

// @ID						CreateServiceAccountOAuthClient
// @Summary				Create an OAuth client for a service account
// @Description			Creates a confidential OAuth 2.0 client bound to the service account. The client authenticates to the token endpoint with client_secret_basic and the client_credentials grant; create a secret separately.
// @Param					account_id	path	string									true	"service account ID"
// @Param					req			body	CreateServiceAccountOAuthClientRequest	true	"Input"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.OAuthClient
// @Router					/v1/service-accounts/{account_id}/oauth-clients [POST]
func (s *service) CreateServiceAccountOAuthClient(ctx *gin.Context) {
	acct, orgID, err := s.orgServiceAccountForOAuth(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateServiceAccountOAuthClientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	client, err := oauthclients.CreateConfidentialClient(ctx.Request.Context(), s.db, orgID, acct.ID, req.Name)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create oauth client: %w", err))
		return
	}

	client.TokenEndpoint = oauthclients.TokenEndpoint(s.cfg)
	ctx.JSON(http.StatusCreated, client)
}

// @ID						ListServiceAccountOAuthClients
// @Summary				List OAuth clients for a service account
// @Param					account_id	path	string	true	"service account ID"
// @Tags					accounts
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.OAuthClient
// @Router					/v1/service-accounts/{account_id}/oauth-clients [GET]
func (s *service) ListServiceAccountOAuthClients(ctx *gin.Context) {
	acct, orgID, err := s.orgServiceAccountForOAuth(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	clients, err := oauthclients.ListAccountClients(ctx.Request.Context(), s.db, orgID, acct.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	for i := range clients {
		clients[i].TokenEndpoint = oauthclients.TokenEndpoint(s.cfg)
	}
	ctx.JSON(http.StatusOK, clients)
}

// @ID						DeleteServiceAccountOAuthClient
// @Summary				Delete an OAuth client and all of its secrets
// @Param					account_id	path	string	true	"service account ID"
// @Param					client_id	path	string	true	"OAuth client ID"
// @Tags					accounts
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				204
// @Router					/v1/service-accounts/{account_id}/oauth-clients/{client_id} [DELETE]
func (s *service) DeleteServiceAccountOAuthClient(ctx *gin.Context) {
	acct, orgID, err := s.orgServiceAccountForOAuth(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	client, err := oauthclients.GetAccountClient(ctx.Request.Context(), s.db, orgID, acct.ID, ctx.Param("client_id"))
	if err != nil {
		ctx.Error(oauthClientError(err))
		return
	}

	if err := oauthclients.DeleteClient(ctx.Request.Context(), s.db, client.ID); err != nil {
		ctx.Error(oauthClientError(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

// @ID						CreateServiceAccountOAuthClientSecret
// @Summary				Create a secret for a service account OAuth client
// @Description			Returns the plaintext client secret once. A client holds at most two active secrets so rotation can overlap: create the replacement, deploy it, then revoke the old one.
// @Param					account_id	path	string							true	"service account ID"
// @Param					client_id	path	string							true	"OAuth client ID"
// @Param					req			body	CreateOAuthClientSecretRequest	true	"Input"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	CreateOAuthClientSecretResponse
// @Router					/v1/service-accounts/{account_id}/oauth-clients/{client_id}/secrets [POST]
func (s *service) CreateServiceAccountOAuthClientSecret(ctx *gin.Context) {
	acct, orgID, err := s.orgServiceAccountForOAuth(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateOAuthClientSecretRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	client, err := oauthclients.GetAccountClient(ctx.Request.Context(), s.db, orgID, acct.ID, ctx.Param("client_id"))
	if err != nil {
		ctx.Error(oauthClientError(err))
		return
	}

	secret, plaintext, err := oauthclients.CreateSecret(ctx.Request.Context(), s.db, client.ID, req.ExpiresAt)
	if err != nil {
		ctx.Error(oauthClientError(err))
		return
	}

	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(http.StatusCreated, CreateOAuthClientSecretResponse{
		Secret:       *secret,
		ClientID:     client.ID,
		ClientSecret: plaintext,
	})
}

// @ID						DeleteServiceAccountOAuthClientSecret
// @Summary				Revoke a service account OAuth client secret
// @Param					account_id	path	string	true	"service account ID"
// @Param					client_id	path	string	true	"OAuth client ID"
// @Param					secret_id	path	string	true	"secret ID"
// @Tags					accounts
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				204
// @Router					/v1/service-accounts/{account_id}/oauth-clients/{client_id}/secrets/{secret_id} [DELETE]
func (s *service) DeleteServiceAccountOAuthClientSecret(ctx *gin.Context) {
	acct, orgID, err := s.orgServiceAccountForOAuth(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	client, err := oauthclients.GetAccountClient(ctx.Request.Context(), s.db, orgID, acct.ID, ctx.Param("client_id"))
	if err != nil {
		ctx.Error(oauthClientError(err))
		return
	}

	if err := oauthclients.RevokeSecret(ctx.Request.Context(), s.db, client.ID, ctx.Param("secret_id")); err != nil {
		ctx.Error(oauthClientError(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}
