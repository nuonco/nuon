package activities

import (
	"context"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/registryauth"
)

// GetACRAccessTokenRequest names the credential rather than carrying it.
//
// ClientSecretName and ClientCertificateName are AppSecret names, resolved
// against the DB inside the activity. Activity inputs are recorded in Temporal
// history, so passing the vendor's long-lived secret here would persist it;
// passing a name persists nothing. The minted refresh token does land in
// history as the activity result, but it expires in 60 minutes — the same
// exposure GAR tokens already have.
type GetACRAccessTokenRequest struct {
	// ComponentID rather than an app ID: it is present on every build plan
	// without depending on a preload, and the activity has a DB handle anyway.
	ComponentID string
	LoginServer string
	TenantID    string
	ClientID    string

	ClientSecretName      string
	ClientCertificateName string
}

type ACRAccessToken struct {
	Username string
	Password string
}

// GetACRAccessToken mints a registry refresh token for a vendor-owned ACR.
//
// Like GetGARAccessToken this lives in the shared activity set because both the
// components namespace (pulling a source image for a build) and the installs
// namespace (resolving a sandbox artifact) schedule it.
//
// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) GetACRAccessToken(ctx context.Context, req *GetACRAccessTokenRequest) (*ACRAccessToken, error) {
	token, err := a.registryAuth.ACR(ctx, temporalzap.GetActivityLogger(ctx), registryauth.ACRRequest{
		ComponentID:           req.ComponentID,
		LoginServer:           req.LoginServer,
		TenantID:              req.TenantID,
		ClientID:              req.ClientID,
		ClientSecretName:      req.ClientSecretName,
		ClientCertificateName: req.ClientCertificateName,
	})
	if err != nil {
		return nil, err
	}

	return &ACRAccessToken{
		Username: token.Username,
		Password: token.Password,
	}, nil
}
