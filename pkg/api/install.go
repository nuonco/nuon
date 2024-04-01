package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type Install struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *client) ListInstalls(ctx context.Context, typ string) ([]Install, error) {
	endpoint := "/v1/installs?type=" + typ
	byts, err := c.execGetRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to execute post request: %w", err)
	}

	var response []Install
	if err := json.Unmarshal(byts, &response); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return response, nil
}

func (c *client) ForgetInstall(ctx context.Context, installID string) error {
	endpoint := fmt.Sprintf("/v1/installs/%s/admin-forget", installID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to forget install: %v", response)
	}

	return nil
}

func (c *client) ReprovisionInstall(ctx context.Context, installID string) error {
	endpoint := fmt.Sprintf("/v1/installs/%s/admin-reprovision", installID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to reprovision install: %v", response)
	}

	return nil
}

func (c *client) RestartInstall(ctx context.Context, installID string) error {
	endpoint := fmt.Sprintf("/v1/installs/%s/admin-restart", installID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to restart install: %v", response)
	}

	return nil
}

func (c *client) UpdateInstallSandbox(ctx context.Context, installID string) error {
	endpoint := fmt.Sprintf("/v1/installs/%s/admin-update-sandbox", installID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to update install sandbox: %v", response)
	}

	return nil
}

func (c *client) DeprovisionInstall(ctx context.Context, installID string) error {
	endpoint := fmt.Sprintf("/v1/installs/%s/admin-deprovision", installID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to deprovision: %v", response)
	}

	return nil
}

func (c *client) DeleteInstall(ctx context.Context, installID string) error {
	endpoint := fmt.Sprintf("/v1/installs/%s/admin-delete", installID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to delete: %v", response)
	}

	return nil
}
