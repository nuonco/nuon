package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func (a *assumer) getGithubOIDCToken(ctx context.Context) (string, error) {
	idTokenRequestToken := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	idTokenRequestURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")

	if idTokenRequestToken == "" || idTokenRequestURL == "" {
		return "", fmt.Errorf("GitHub OIDC environment variables not set")
	}
	url := idTokenRequestURL + "&audience=sts.amazonaws.com"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+idTokenRequestToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get OIDC token: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenResponse struct {
		Value string `json:"value"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResponse.Value == "" {
		return "", fmt.Errorf("empty token received")
	}

	return tokenResponse.Value, nil
}
