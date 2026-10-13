package acr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/azure/credentials"
)

// DeleteRepository deletes a repository and all of its images from the registry.
// A repository that does not exist is not an error. A nil cfg uses the ambient identity.
func DeleteRepository(ctx context.Context, cfg *credentials.Config, acrService, repository string, logger *zap.Logger) error {
	refreshToken, err := GetRepositoryToken(ctx, cfg, acrService, logger)
	if err != nil {
		return fmt.Errorf("unable to get refresh token: %w", err)
	}

	accessToken, err := getAccessToken(ctx, acrService, "repository:"+repository+":delete", refreshToken)
	if err != nil {
		return fmt.Errorf("unable to get access token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("https://%s/acr/v1/%s", acrService, repository), nil)
	if err != nil {
		return fmt.Errorf("unable to create delete request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to delete repository: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted, http.StatusNotFound:
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unable to delete repository: unexpected status %d: %s", resp.StatusCode, string(body))
}

func getAccessToken(ctx context.Context, acrService, scope, refreshToken string) (string, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"service":       {acrService},
		"scope":         {scope},
		"refresh_token": {refreshToken},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://%s/oauth2/token", acrService), strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("unable to parse token response: %w", err)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("token response missing access_token")
	}

	return out.AccessToken, nil
}
