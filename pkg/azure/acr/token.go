package acr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/azure/credentials"
)

const (
	DefaultACRUsername string = "00000000-0000-0000-0000-000000000000"
)

func GetRepositoryToken(ctx context.Context, cfg *credentials.Config, acrService string, logger *zap.Logger) (string, error) {
	credential, err := credentials.Fetch(ctx, cfg, logger)
	if err != nil {
		return "", fmt.Errorf("unable to get credential: %w", err)
	}

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

	// why: The exchange must name the tenant that owns the registry. Normally that
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

	return token, nil
}
