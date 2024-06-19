package api

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *client) handleStatusResponse(byts []byte) error {
	var resp map[string]string
	if err := json.Unmarshal(byts, &resp); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}

	if resp["status"] != "ok" {
		return fmt.Errorf("invalid response: %v", resp)
	}

	return nil
}

type ProvisionCanaryRequest struct {
	SandboxMode bool `json:"sandbox_mode"`
}

func (c *client) ProvisionCanary(ctx context.Context, sandboxMode bool) error {
	endpoint := "/v1/general/provision-canary"
	byts, err := c.execPostRequest(ctx, endpoint, ProvisionCanaryRequest{
		SandboxMode: sandboxMode,
	})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	return c.handleStatusResponse(byts)
}

type DeprovisionCanaryRequest struct {
	CanaryID string `json:"canary_id"`
}

func (c *client) DeprovisionCanary(ctx context.Context, canaryID string) error {
	endpoint := "/v1/general/deprovision-canary"
	byts, err := c.execPostRequest(ctx, endpoint, DeprovisionCanaryRequest{
		CanaryID: canaryID,
	})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	return c.handleStatusResponse(byts)
}

type StartCanaryCron struct{}

func (c *client) StartCanaryCron(ctx context.Context) error {
	endpoint := "/v1/general/start-canary-cron"
	byts, err := c.execPostRequest(ctx, endpoint, StartCanaryCron{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	return c.handleStatusResponse(byts)
}

type StopCanaryCron struct{}

func (c *client) StopCanaryCron(ctx context.Context) error {
	endpoint := "/v1/general/stop-canary-cron"
	byts, err := c.execPostRequest(ctx, endpoint, StartCanaryCron{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	return c.handleStatusResponse(byts)
}

type CreateCanaryUserRequest struct {
	CanaryID string `json:"canary_id"`
}

type CreateCanaryUserResponse struct {
	APIToken        string `json:"api_token"`
	GithubInstallID string `json:"github_install_id"`
	Email           string `json:"email"`
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
