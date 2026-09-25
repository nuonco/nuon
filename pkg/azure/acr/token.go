package acr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/azure/credentials"
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
	credential, err := credentials.Fetch(ctx, cfg, logger)
	if err != nil {
		return "", fmt.Errorf("unable to get credential: %w", err)
	}
	tenantID := ""
	if cfg != nil {
		tenantID = cfg.TenantID
	}
	return GetRepositoryTokenWithCredential(ctx, credential, tenantID, acrService)
}

func GetRepositoryTokenWithCredential(ctx context.Context, credential azcore.TokenCredential, tenantID, acrService string) (string, error) {
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

	if tenantID == "" {
		tenantID = claims.TenantID
	}

	formData := url.Values{
		"grant_type":   {"access_token"},
		"service":      {acrService},
		"tenant":       {tenantID},
		"access_token": {aadToken.Token},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://%s/oauth2/exchange", acrService), strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("unable to build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	jsonResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to get credential: %w", err)
	}
	defer jsonResponse.Body.Close()
	if jsonResponse.StatusCode < http.StatusOK || jsonResponse.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("ACR token exchange returned %s", jsonResponse.Status)
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
