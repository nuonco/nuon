package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type App struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *client) ListApps(ctx context.Context) ([]App, error) {
	endpoint := "/v1/apps"
	byts, err := c.execGetRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to execute post request: %w", err)
	}

	var response []App
	if err := json.Unmarshal(byts, &response); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return response, nil
}

func (c *client) ReprovisionApp(ctx context.Context, appID string) error {
	endpoint := fmt.Sprintf("/v1/apps/%s/admin-reprovision", appID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to reprovision app: %v", response)
	}

	return nil
}

func (c *client) RestartApp(ctx context.Context, appID string) error {
	endpoint := fmt.Sprintf("/v1/apps/%s/admin-restart", appID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to restart app: %v", response)
	}

	return nil
}

func (c *client) UpdateAppSandbox(ctx context.Context, appID string) error {
	endpoint := fmt.Sprintf("/v1/apps/%s/admin-update-sandbox", appID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to update app sandbox: %v", response)
	}

	return nil
}
