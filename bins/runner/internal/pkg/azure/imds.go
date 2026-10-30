package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/nuonco/nuon/pkg/azure"
)

const (
	imdsTokenURL = "http://169.254.169.254/metadata/identity/oauth2/token"
)

type imdsTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
	Resource    string `json:"resource"`
	TokenType   string `json:"token_type"`
}

func imdsTokenRequestURL(resource string) string {
	if resource == "" || resource == azure.ManagementAudience {
		return imdsTokenURL + "?api-version=2018-02-01&resource=https://management.azure.com/"
	}
	return imdsTokenURL + "?api-version=2018-02-01&resource=" + url.QueryEscape(resource)
}

// GetIMDSToken fetches a managed identity access token from the Azure IMDS endpoint
// for the given resource. An empty resource requests the ARM audience.
func GetIMDSToken(ctx context.Context, resource string) (string, error) {
	reqURL := imdsTokenRequestURL(resource)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create IMDS request: %w", err)
	}
	req.Header.Set("Metadata", "true")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get Azure IMDS token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Azure IMDS returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read IMDS response: %w", err)
	}

	var tokenResp imdsTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse IMDS response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token from Azure IMDS")
	}

	return tokenResp.AccessToken, nil
}
