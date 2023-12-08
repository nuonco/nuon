package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type CreateIntegrationUserRequest struct{}

type CreateIntegrationUserResponse struct {
	APIToken        string `json:"api_token"`
	GithubInstallID string `json:"github_install_id"`
}

func (c *client) CreateIntegrationUser(ctx context.Context) (*CreateIntegrationUserResponse, error) {
	endpoint := "/v1/general/integration-user"
	byts, err := c.execPostRequest(ctx, endpoint, CreateIntegrationUserResponse{})
	if err != nil {
		return nil, fmt.Errorf("unable to execute post request: %w", err)
	}

	var resp CreateIntegrationUserResponse
	if err := json.Unmarshal(byts, &resp); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return &resp, nil
}

type CreateCanaryUserRequest struct {
	CanaryID string `json:"canary_id"`
}

type CreateCanaryUserResponse struct {
	APIToken        string `json:"api_token"`
	GithubInstallID string `json:"github_install_id"`
}

func (c *client) CreateCanaryUser(ctx context.Context, canaryID string) (*CreateCanaryUserResponse, error) {
	endpoint := "/v1/general/canary-user"
	byts, err := c.execPostRequest(ctx, endpoint, CreateCanaryUserRequest{
		CanaryID: canaryID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute post request: %w", err)
	}

	var resp CreateCanaryUserResponse
	if err := json.Unmarshal(byts, &resp); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return &resp, nil
}
