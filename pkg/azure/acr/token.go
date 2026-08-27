package acr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/nuonco/nuon/pkg/azure/credentials"
	"go.uber.org/zap"
)

const (
	DefaultACRUsername string = "00000000-0000-0000-0000-000000000000"
)

// GetRepositoryToken exchanges an Azure credential for a refresh token that can be used to authenticate with
// the registry. It has a timeout of 60 minutes.
//
// cfg selects the identity: an app registration when it carries one (the only
// way to reach a registry in another tenant), otherwise the ambient identity.
// A nil cfg means ambient.
//
// NOTE: we do this, instead of using the ACR repository client to simplify our dependencies, however, at some point we
// plan on moving this into a package, like we have with `pkg/aws`.
func GetRepositoryToken(ctx context.Context, cfg *credentials.Config, acrService string, logger *zap.Logger) (string, error) {
	// get a credential
	credential, err := credentials.Fetch(ctx, cfg, logger)
	if err != nil {
		return "", fmt.Errorf("unable to get credential: %w", err)
	}

	// use the credentials to get an Entra ID token
	aadToken, err := credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"}},
	)
	if err != nil {
		return "", fmt.Errorf("unable to get credential: %w", err)
	}

	claims, err := parseJWT(aadToken.Token)
	if err != nil {
		return "", fmt.Errorf("unable to parse entra id token for claims: %w", err)
	}

	// The exchange must name the tenant that owns the registry. Normally that
	// is the token's own tenant, but an app registration authenticating into a
	// vendor tenant is configured with it explicitly, so prefer that.
	tenantID := claims.TenantID
	if cfg != nil && cfg.TenantID != "" {
		tenantID = cfg.TenantID
	}

	formData := url.Values{
		"grant_type":   {"access_token"},
		"service":      {acrService},
		"tenant":       {tenantID},
		"access_token": {aadToken.Token},
	}
	jsonResponse, err := http.PostForm(fmt.Sprintf("https://%s/oauth2/exchange", acrService), formData)
	if err != nil {
		return "", fmt.Errorf("unable to get credential: %w", err)
	}
	var response map[string]interface{}
	decoder := json.NewDecoder(jsonResponse.Body)
	if err := decoder.Decode(&response); err != nil {
		return "", fmt.Errorf("unable to parse token response: %w", err)
	}
	rawToken := response["refresh_token"]
	token, ok := rawToken.(string)
	if !ok {
		return "", fmt.Errorf("unable to parse refresh token as string")
	}

	// return the refresh token
	return token, nil
}
