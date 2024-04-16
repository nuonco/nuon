package acr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/powertoolsdev/mono/pkg/azure/credentials"
)

const (
	DefaultACRUsername string = "00000000-0000-0000-0000-000000000000"
)

// getACRRepositorToken exchanges an Azure credential, and creates a refresh token that can be used to authenticate with
// the registry. It has a timeout of 60 minutes.
//
// NOTE: we do this, instead of using the ACR repository client to simplify our dependencies, however, at some point we
// plan on moving this into a package, like we have with `pkg/aws`.
func GetRepositoryToken(ctx context.Context, cfg *credentials.Config, acrService string) (string, error) {
	credential, err := credentials.Fetch(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("unable to get credential: %w", err)
	}

	aadToken, err := credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"}},
	)
	if err != nil {
		return "", fmt.Errorf("unable to get credential: %w", err)
	}

	formData := url.Values{
		"grant_type":   {"access_token"},
		"service":      {acrService},
		"tenant":       {cfg.ServicePrincipal.SubscriptionTenantID},
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
