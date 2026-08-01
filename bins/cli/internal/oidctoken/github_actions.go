package oidctoken

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	ghaRequestURLEnvVar   = "ACTIONS_ID_TOKEN_REQUEST_URL"
	ghaRequestTokenEnvVar = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
)

func githubActionsAvailable() bool {
	return os.Getenv(ghaRequestURLEnvVar) != "" && os.Getenv(ghaRequestTokenEnvVar) != ""
}

// fetchGitHubActionsToken requests an OIDC ID token from the GitHub Actions
// runtime. Requires the workflow to grant `permissions: id-token: write`.
func fetchGitHubActionsToken(ctx context.Context, audience string) (string, error) {
	requestURL := os.Getenv(ghaRequestURLEnvVar)
	requestToken := os.Getenv(ghaRequestTokenEnvVar)

	if audience != "" {
		parsed, err := url.Parse(requestURL)
		if err != nil {
			return "", fmt.Errorf("invalid %s: %w", ghaRequestURLEnvVar, err)
		}
		q := parsed.Query()
		q.Set("audience", audience)
		parsed.RawQuery = q.Encode()
		requestURL = parsed.String()
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create ID token request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+requestToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub Actions ID token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("unable to read ID token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub Actions ID token request returned %d", resp.StatusCode)
	}

	var payload struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("unable to parse ID token response: %w", err)
	}
	if payload.Value == "" {
		return "", fmt.Errorf("GitHub Actions ID token response was empty")
	}

	return payload.Value, nil
}
