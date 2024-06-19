package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type CreateSeedUserRequest struct{}

type CreateSeedUserResponse struct {
	APIToken        string `json:"api_token"`
	GithubInstallID string `json:"github_install_id"`
	Email           string `json:"email"`
}

func (c *client) CreateSeedUser(ctx context.Context) (*CreateSeedUserResponse, error) {
	endpoint := "/v1/general/seed-user"
	byts, err := c.execPostRequest(ctx, endpoint, CreateSeedUserResponse{})
	if err != nil {
		return nil, fmt.Errorf("unable to execute post request: %w", err)
	}

	var resp CreateSeedUserResponse
	if err := json.Unmarshal(byts, &resp); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return &resp, nil
}
